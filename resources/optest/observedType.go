package optest

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/resources"
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
	observer   *ObservedType
	seenEvents eventCounter
}

func (a *TypeAspect) update() {
	a.seenEvents++
}

func (a *TypeAspect) Fork() *ChangePoint {
	return &ChangePoint{
		aspect: a,
		origin: a.seenEvents,
	}
}

func (a *TypeAspect) events() eventCounter {
	return a.seenEvents
}

func (a *TypeAspect) consumeEvent(ctx context.Context) error {
	return a.observer.consumeEvent(ctx)
}
