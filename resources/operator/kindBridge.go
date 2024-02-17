package operator

import (
	"context"
	"errors"
	"github.com/meschbach/go-junk-bucket/pkg/fx"
	"github.com/meschbach/plaid/resources"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type KindBridgeState struct {
	//OnResourceChange schedules an update
	OnResourceChange func(ctx context.Context, which resources.Meta) error
}

type KindBridgeBinding[Spec any, Status any, R any] interface {
	Create(ctx context.Context, which resources.Meta, spec Spec, bridge *KindBridgeState) (*R, Status, error)
	Update(ctx context.Context, which resources.Meta, rt *R, s Spec) (Status, error)
	Delete(ctx context.Context, which resources.Meta, rt *R) error
}

// KindBridge bridges resource management of a specific kind
type KindBridge[Spec any, Status any, R any] struct {
	//todo: reflect on scoping and Setup -- does not exist until setting up
	//changeFeed allows injection of updating in a control loop
	changeFeed chan resources.ResourceChanged
	store      *resources.Client
	observer   *resources.ClientWatcher
	kind       resources.Type
	binding    KindBridgeBinding[Spec, Status, R]
	mapping    resources.MetaContainer[R]
}

func NewKindBridge[Spec any, Status any, Runtime any](resourceType resources.Type, ops KindBridgeBinding[Spec, Status, Runtime]) *KindBridge[Spec, Status, Runtime] {
	return &KindBridge[Spec, Status, Runtime]{
		kind:    resourceType,
		binding: ops,
	}
}

func (k *KindBridge[P, T, R]) Setup(parentCtx context.Context, r *resources.Client) (chan resources.ResourceChanged, error) {
	ctx, span := tracing.Start(parentCtx, "KindBridge["+k.kind.String()+"].Setup")
	defer span.End()

	k.store = r
	//todo: have a client pass this in
	listener, err := r.Watcher(ctx)
	if err != nil {
		return nil, err
	}
	changes := listener.Feed
	k.observer = listener
	k.changeFeed = changes
	_, err = listener.OnType(ctx, k.kind, func(ctx context.Context, changed resources.ResourceChanged) error {
		return k.digestChange(ctx, changed)
	})
	if err != nil {
		return nil, err
	}

	//find matching
	rescanError := k.Rescan(ctx)
	if rescanError != nil {
		close(changes)
	}
	return changes, rescanError
}

// Rescan reviews the state of the store and attempts to synchronize the resources
func (k *KindBridge[Spec, Status, R]) Rescan(parentCtx context.Context) error {
	ctx, span := tracing.Start(parentCtx, "KindBridge["+k.kind.String()+"]#Rescan ")
	defer span.End()
	found, err := k.store.List(ctx, k.kind)
	if err != nil {
		return err
	}

	deletedMetas := k.mapping.AllMetas()
	var problems []error
	for _, ref := range found {
		_, hasState := k.mapping.Find(ref)
		if hasState {
			span.AddEvent("sync", trace.WithAttributes(ref.AsTraceAttribute("ref")...))
			deletedMetas = fx.Filter(deletedMetas, func(e resources.Meta) bool {
				return !ref.EqualsMeta(e)
			})
			if err := k.updated(ctx, k.store, ref); err != nil {
				span.SetStatus(codes.Error, "update failed")
				problems = append(problems, err)
			}
		} else {
			span.AddEvent("create", trace.WithAttributes(ref.AsTraceAttribute("ref")...))
			if err := k.create(ctx, k.store, ref); err != nil {
				span.SetStatus(codes.Error, "create failed")
				problems = append(problems, err)
			}
		}
	}

	//delete all remaining
	for _, ref := range deletedMetas {
		span.AddEvent("delete missing", trace.WithAttributes(attribute.Stringer("resource", ref)))
		if err := k.delete(ctx, k.store, ref); err != nil {
			span.SetStatus(codes.Error, "delete failed")
			problems = append(problems, err)
		}
	}
	return errors.Join(problems...)
}

func (k *KindBridge[Spec, Status, R]) create(parentCtx context.Context, r *resources.Client, which resources.Meta) error {
	ctx, span := tracing.Start(parentCtx, "KindBridge["+which.Type.String()+"]#create", trace.WithAttributes(attribute.Stringer("which", which)))
	defer span.End()

	//todo: figure out how we can get multiple create events for the same object
	_, found := k.mapping.Find(which)
	if found {
		return k.updated(ctx, r, which)
	}

	var spec Spec
	exists, err := r.Get(ctx, which, &spec)
	if err != nil {
		span.SetStatus(codes.Error, "unable to locate spec")
		return err
	}
	if !exists {
		span.AddEvent("missing")
		return nil
	}

	clientCallbacks := &KindBridgeState{
		OnResourceChange: func(parent context.Context, which resources.Meta) error {
			if !which.Type.Equals(k.kind) {
				panic("update on unmanaged type")
			}
			ctx, span := tracing.Start(parent, "KindBridge["+k.kind.Kind+"].bridge.OnResourceChange", trace.WithAttributes(which.AsTraceAttribute("which")...))
			defer span.End()
			select {
			case k.changeFeed <- resources.ResourceChanged{
				Which:     which,
				Operation: resources.UpdatedEvent,
				Tracing:   trace.LinkFromContext(ctx),
			}:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	}
	runtimeInstance, status, err := k.binding.Create(ctx, which, spec, clientCallbacks)
	if err != nil {
		return err
	} //todo: should just record an oerror on the resource

	k.mapping.Upsert(which, runtimeInstance) //todo: handle conflicts and graceful shutdown

	statusExists, err := r.UpdateStatus(ctx, which, status)
	if err != nil {
		span.SetStatus(codes.Error, "failed to update status")
		return err
	}
	if !statusExists {
		span.SetStatus(codes.Error, "resource went missing")
		//todo: delete
	}
	span.SetStatus(codes.Ok, "success")
	return nil
}

func (k *KindBridge[Spec, Status, R]) updated(parentCtx context.Context, r *resources.Client, changed resources.Meta) error {
	ctx, span := tracing.Start(parentCtx, "KindBridge["+changed.Type.String()+"]#updated ", trace.WithAttributes(changed.AsTraceAttribute("which")...))
	defer span.End()

	var s Spec
	exists, err := r.Get(ctx, changed, &s)
	if err != nil {
		return err
	}
	if !exists {
		span.AddEvent("spec-missing")
		return nil
	}
	rt, has := k.mapping.Find(changed)
	if has {
		span.AddEvent("updating")
		status, err := k.binding.Update(ctx, changed, rt, s)
		if err != nil {
			return err
		}
		exists, err := r.UpdateStatus(ctx, changed, status)
		if err != nil {
			return err
		}
		if !exists { //todo: ensure ignoring and waiting for deletion is appropriate
			return nil
		}
		return nil
	} else {
		span.AddEvent("creating")
		return k.create(ctx, r, changed)
	}
}

func (k *KindBridge[Spec, Status, R]) delete(parentCtx context.Context, r *resources.Client, changed resources.Meta) error {
	state, has := k.mapping.Delete(changed)
	if !has { //for whatever reason we don't have state associated with this resource, ignore it
		return nil
	}
	problem := k.binding.Delete(parentCtx, changed, state)
	return problem
}

func (k *KindBridge[Spec, Status, R]) Dispatch(parent context.Context, r *resources.Client, changed resources.ResourceChanged) error {
	return k.observer.Digest(parent, changed)
}

func (k *KindBridge[Spec, Status, R]) digestChange(parent context.Context, changed resources.ResourceChanged) error {
	ctx, span := tracing.Start(parent, "KindBridge["+k.kind.Kind+"]#digestChange")
	defer span.End()

	switch changed.Operation {
	case resources.CreatedEvent:
		return k.create(ctx, k.store, changed.Which)
	case resources.DeletedEvent:
		return k.delete(ctx, k.store, changed.Which)
	case resources.UpdatedEvent:
		return k.updated(ctx, k.store, changed.Which)
	default:
		return nil
	}
}

func (k *KindBridge[Spec, Status, R]) All() []*R {
	return k.mapping.AllValues()
}

// ConsumeEvent will wait on an event to be processed or for the parent context to be cancelled.  Generally useful in
// testing code when one knows there are specific events to be processed.
func (k *KindBridge[Spec, Status, R]) ConsumeEvent(parentContext context.Context) error {
	select {
	case <-parentContext.Done():
		return parentContext.Err()
	case e := <-k.changeFeed:
		return k.observer.Digest(parentContext, e)
	}
}
