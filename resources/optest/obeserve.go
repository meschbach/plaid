package optest

import (
	"context"
	"github.com/meschbach/plaid/resources"
	"github.com/stretchr/testify/require"
	"testing"
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
	observer      *ObservedResource
	changedEvents eventCounter
}

func (a *ResourceAspect) update() {
	a.changedEvents++
}

func (a *ResourceAspect) Fork() *ChangePoint {
	return &ChangePoint{
		aspect: a,
		origin: a.changedEvents,
	}
}

func (a *ResourceAspect) events() eventCounter {
	return a.changedEvents
}

func (a *ResourceAspect) consumeEvent(ctx context.Context) error {
	return a.observer.consumeEvent(ctx)
}

type eventCounter uint64
type Aspect interface {
	consumeEvent(ctx context.Context) error
	events() eventCounter
}

type ChangePoint struct {
	aspect Aspect
	origin eventCounter
}

func (r *ChangePoint) Wait(t *testing.T, ctx context.Context) {
	for r.origin >= r.aspect.events() {
		err := r.aspect.consumeEvent(ctx)
		if err != nil {
			require.NoError(t, err)
			return
		}
	}
}

func (r *ChangePoint) WaitFor(t *testing.T, ctx context.Context, satisfied func(ctx context.Context) bool) {
	for !satisfied(ctx) {
		r.Wait(t, ctx)
	}
}
