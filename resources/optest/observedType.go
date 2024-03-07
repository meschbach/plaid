package optest

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/resources"
	"github.com/stretchr/testify/require"
	"testing"
)

// todo: a lot of the aspect things could probably be reused/DRYed
type ObservedType struct {
	system       *System
	token        resources.WatchToken
	AnyEvent     *TypeAspect
	Create       *TypeAspect
	Delete       *TypeAspect
	Update       *TypeAspect
	UpdateStatus *TypeAspect
}

func (o *ObservedType) onResourceEvent(ctx context.Context, event resources.ResourceChanged) error {
	o.AnyEvent.update()
	switch event.Operation {
	case resources.CreatedEvent:
		o.Create.update()
	case resources.DeletedEvent:
		o.Delete.update()
	case resources.UpdatedEvent:
		o.Update.update()
	case resources.StatusUpdated:
		o.UpdateStatus.update()
	default:
		panic(fmt.Sprintf("Unknown event type %s", event.Operation))
	}
	return nil
}

func (o *ObservedType) consumeEvent(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-o.system.observer.Events():
		err := o.system.observer.Digest(ctx, e)
		return err
	}
}

type TypeAspect struct {
	observer *ObservedType
	events   uint64
}

func (a *TypeAspect) update() {
	a.events++
}

func (a *TypeAspect) Fork() *TypeChangePoint {
	return &TypeChangePoint{
		aspect: a,
		origin: a.events,
	}
}

func (a *TypeAspect) consumeEvent(ctx context.Context) error {
	return a.observer.consumeEvent(ctx)
}

type TypeChangePoint struct {
	aspect *TypeAspect
	//origin is the event the change point was created at
	origin uint64
}

func (r *TypeChangePoint) Wait(t *testing.T, ctx context.Context) {
	t.Helper()
	for r.origin >= r.aspect.events {
		err := r.aspect.consumeEvent(ctx)
		if err != nil {
			require.NoError(t, err)
			return
		}
	}
}

func (r *TypeChangePoint) WaitFor(t *testing.T, ctx context.Context, satisfied func(ctx context.Context) bool) {
	for !satisfied(ctx) {
		r.Wait(t, ctx)
	}
}
