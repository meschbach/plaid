package optest

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/resources"
	"github.com/stretchr/testify/require"
)

func (s *System) ObserveType(ctx context.Context, kind resources.Type) *ObservedType {
	observer, created := s.typeObservers.GetOrCreate(kind, func() *ObservedType {
		o := &ObservedType{
			system: s,
		}
		o.AnyEvent = &Aspect{observer: o}
		o.Create = &Aspect{observer: o}
		o.Delete = &Aspect{observer: o}
		o.Update = &Aspect{observer: o}
		o.UpdateStatus = &Aspect{observer: o}
		return o
	})
	if created {
		token, err := s.observer.OnType(ctx, kind, observer.onResourceEvent)
		require.NoError(s.t, err)
		observer.token = token
	}
	return observer
}

type ObservedType struct {
	system       *System
	token        resources.WatchToken
	AnyEvent     *Aspect
	Create       *Aspect
	Delete       *Aspect
	Update       *Aspect
	UpdateStatus *Aspect
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
