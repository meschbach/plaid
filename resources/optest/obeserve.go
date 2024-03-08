package optest

import (
	"context"
	"github.com/meschbach/plaid/resources"
	"github.com/stretchr/testify/require"
)

type ObservedResource struct {
	system   *System
	token    resources.WatchToken
	AnyEvent *ResourceAspect
	Spec     *ResourceAspect
	Status   *ResourceAspect
}

func (o *ObservedResource) onResourceEvent(ctx context.Context, event resources.ResourceChanged) error {
	o.AnyEvent.update()
	switch event.Operation {
	case resources.UpdatedEvent:
		o.Spec.update()
	case resources.StatusUpdated:
		o.Status.update()
	default:
	}
	return nil
}

func (o *ObservedResource) consumeEvent(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-o.system.observer.Events():
		return o.system.observer.Digest(ctx, e)
	}
}

type ResourceAspect struct {
	observer *ObservedResource
	events   uint64
}

func (a *ResourceAspect) update() {
	a.events++
}

func (a *ResourceAspect) Fork() *ResourceChangePoint {
	return &ResourceChangePoint{
		aspect: a,
		origin: a.events,
	}
}

func (a *ResourceAspect) consumeEvent(ctx context.Context) error {
	return a.observer.consumeEvent(ctx)
}

type ResourceChangePoint struct {
	aspect *ResourceAspect
	//origin is the event the change point was created at
	origin uint64
}

func (r *ResourceChangePoint) Wait(ctx context.Context) {
	for r.origin >= r.aspect.events {
		require.NoError(r.aspect.observer.system.t, r.aspect.consumeEvent(ctx))
	}
}

func (r *ResourceChangePoint) WaitFor(ctx context.Context, satisfied func(ctx context.Context) bool) {
	for !satisfied(ctx) {
		r.Wait(ctx)
	}
}
