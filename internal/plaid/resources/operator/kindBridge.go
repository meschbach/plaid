package operator

import (
	"context"
	"fmt"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
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
	ctx, span := tracing.Start(parentCtx, "KindBridge.Setup "+k.kind.String())
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
	matching, err := r.List(ctx, k.kind)
	if err != nil {
		close(changes)
		return nil, err
	}

	for _, existing := range matching {
		if err := k.create(ctx, r, existing); err != nil {
			close(changes)
			return nil, err
		}
	}

	// channel to be told of
	return changes, nil
}

func (k *KindBridge[Spec, Status, R]) create(parentCtx context.Context, r *resources.Client, which resources.Meta) error {
	ctx, span := tracing.Start(parentCtx, "KindBridge#create", trace.WithAttributes(attribute.Stringer("which", which)))
	defer span.End()

	span.SetName("KindBridge#create " + which.Type.String())

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
		OnResourceChange: func(ctx context.Context, which resources.Meta) error {
			if !which.Type.Equals(k.kind) {
				panic("update on unmanaged type")
			}
			span.AddEvent("KindBridgeState#OnResourceChange", trace.WithAttributes(attribute.Stringer("which", which)))
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
	ctx, span := tracing.Start(parentCtx, "KindBridge#updated", trace.WithAttributes(attribute.Stringer("which", changed)))
	defer span.End()

	span.SetName("KindBridge#updated " + changed.Type.String())

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
			fmt.Println("resource went missing during update")
			return nil
		}
		return nil
	} else {
		span.AddEvent("creating")
		return k.create(ctx, r, changed)
	}
}

func (k *KindBridge[Spec, Status, R]) Dispatch(parent context.Context, r *resources.Client, changed resources.ResourceChanged) error {
	return k.observer.Digest(parent, changed)
}

func (k *KindBridge[Spec, Status, R]) digestChange(parent context.Context, changed resources.ResourceChanged) error {
	ctx, span := tracing.Start(parent, "KindBridge.")
	defer span.End()

	span.SetName("KindBridge." + changed.Operation.String())

	switch changed.Operation {
	case resources.CreatedEvent:
		return k.create(ctx, k.store, changed.Which)
	case resources.DeletedEvent: //todo: properly clean up
		return nil
	case resources.UpdatedEvent:
		return k.updated(ctx, k.store, changed.Which)
	default:
		return nil
	}
}

func (k *KindBridge[Spec, Status, R]) All() []*R {
	return k.mapping.AllValues()
}
