package kit

import (
	"context"
	"errors"
	"fmt"
	"github.com/meschbach/go-junk-bucket/pkg/fx"
	"github.com/meschbach/plaid/resources"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type state[State any] struct {
	runtimeState *State
}

type Kit[Spec any, Status any, State any] struct {
	store    resources.Storage
	observer resources.Watcher
	kind     resources.Type
	bridge   Operations[Spec, Status, State]
	mapping  *resources.MetaContainer[state[State]]
}

func New[Spec any, Status any, State any](manage resources.Storage, observer resources.Watcher, kind resources.Type, bridge Operations[Spec, Status, State]) *Kit[Spec, Status, State] {
	kit := &Kit[Spec, Status, State]{
		store:    manage,
		observer: observer,
		kind:     kind,
		bridge:   bridge,
		mapping:  resources.NewMetaContainer[state[State]](),
	}
	return kit
}

func (k *Kit[Spec, Status, State]) Setup(parentCtx context.Context) error {
	ctx, span := tracer.Start(parentCtx, "Kit["+k.kind.String()+"].Setup")
	defer span.End()

	_, err := k.observer.OnType(ctx, k.kind, k.digestChange)
	if err != nil {
		span.SetStatus(codes.Error, "failed to observe type")
		return err
	}

	return k.Rescan(ctx)
}

func (k *Kit[Spec, Status, State]) digestChange(parent context.Context, changed resources.ResourceChanged) error {
	ctx, span := tracer.Start(parent, "Kit["+k.kind.Kind+"]#digestChange")
	defer span.End()

	switch changed.Operation {
	case resources.CreatedEvent:
		return k.create(ctx, changed.Which)
	case resources.DeletedEvent:
		return k.delete(ctx, changed.Which)
	case resources.UpdatedEvent:
		return k.updated(ctx, changed.Which)
	default:
		return nil
	}
}

func (k *Kit[Spec, Status, State]) Rescan(parentCtx context.Context) error {
	ctx, span := tracer.Start(parentCtx, "Kit["+k.kind.String()+"]#Rescan ")
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
			if err := k.updated(ctx, ref); err != nil {
				span.SetStatus(codes.Error, "update failed")
				problems = append(problems, err)
			}
		} else {
			span.AddEvent("create", trace.WithAttributes(ref.AsTraceAttribute("ref")...))
			if err := k.create(ctx, ref); err != nil {
				span.SetStatus(codes.Error, "create failed")
				problems = append(problems, err)
			}
		}
	}

	//delete all remaining
	for _, ref := range deletedMetas {
		span.AddEvent("delete missing", trace.WithAttributes(attribute.Stringer("resource", ref)))
		if err := k.delete(ctx, ref); err != nil {
			span.SetStatus(codes.Error, "delete failed")
			problems = append(problems, err)
		}
	}
	return errors.Join(problems...)
}

func (k *Kit[Spec, Status, State]) create(parentCtx context.Context, which resources.Meta) error {
	ctx, span := tracer.Start(parentCtx, "Kit["+which.Type.String()+"]#create", trace.WithAttributes(attribute.Stringer("which", which)))
	defer span.End()

	//todo: figure out how we can get multiple create events for the same object
	_, found := k.mapping.Find(which)
	if found {
		return k.updated(ctx, which)
	}

	var spec Spec
	exists, err := k.store.Get(ctx, which, &spec)
	if err != nil {
		span.SetStatus(codes.Error, "unable to locate spec")
		return err
	}
	if !exists {
		span.AddEvent("missing")
		return nil
	}

	manager := &observerInjectedManager{
		target:   which,
		observer: k.observer,
	}
	runtimeState, err := k.bridge.Create(ctx, which, spec, manager)
	if err != nil {
		return err
	}
	kitState := &state[State]{
		runtimeState: runtimeState,
	}
	k.mapping.Upsert(which, kitState) //todo: get or create above should setup the inital state

	return k.updateStatus(ctx, span, which, kitState)
}

func (k *Kit[Spec, Status, State]) updated(parentCtx context.Context, changed resources.Meta) error {
	fmt.Println("kit update")
	ctx, span := tracer.Start(parentCtx, "Kit["+changed.Type.String()+"]#updated ", trace.WithAttributes(changed.AsTraceAttribute("which")...))
	defer span.End()

	var s Spec
	exists, err := k.store.Get(ctx, changed, &s)
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
		err := k.bridge.Update(ctx, changed, rt.runtimeState, s)
		if err != nil {
			return err
		}
		return k.updateStatus(ctx, span, changed, rt)
	} else {
		span.AddEvent("creating")
		return k.create(ctx, changed)
	}
}

func (k *Kit[Spec, Status, State]) updateStatus(ctx context.Context, span trace.Span, which resources.Meta, kitState *state[State]) error {
	status := k.bridge.Status(ctx, kitState.runtimeState)
	statusExists, err := k.store.UpdateStatus(ctx, which, status)
	if err != nil {
		span.SetStatus(codes.Error, "failed")
		return err
	}
	if !statusExists {
		span.SetStatus(codes.Error, "missing")
		//todo: delete?  we should probably wait until we get the event.
	}
	return nil
}

func (k *Kit[Spec, Status, State]) delete(parentCtx context.Context, changed resources.Meta) error {
	state, has := k.mapping.Delete(changed)
	if !has { //for whatever reason we don't have state associated with this resource, ignore it
		return nil
	}
	problem := k.bridge.Delete(parentCtx, changed, state.runtimeState)
	return problem
}
