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
	AnyEvent *Aspect
	Spec     *Aspect
	Status   *Aspect
}

func (s *System) Observe(ctx context.Context, ref resources.Meta) *ObservedResource {
	observer, created := s.observers.GetOrCreate(ref, func() *ObservedResource {
		o := &ObservedResource{
			system: s,
		}
		o.AnyEvent = &Aspect{observer: o}
		o.Spec = &Aspect{observer: o}
		o.Status = &Aspect{observer: o}
		return o
	})
	if created {
		token, err := s.observer.OnResource(ctx, ref, observer.onResourceEvent)
		require.NoError(s.t, err)
		observer.token = token
	}
	return observer
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

type eventCounter uint64
type Observatory interface {
	consumeEvent(ctx context.Context) error
}

type Aspect struct {
	observer  Observatory
	seenCount eventCounter
}

func (a *Aspect) events() eventCounter {
	return a.seenCount
}

func (a *Aspect) consumeEvent(ctx context.Context) error {
	return a.observer.consumeEvent(ctx)
}

func (a *Aspect) update() {
	a.seenCount++
}

func (a *Aspect) Fork() *ChangePoint {
	return &ChangePoint{
		aspect: a,
		origin: a.seenCount,
	}
}

type ChangePoint struct {
	aspect *Aspect
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
