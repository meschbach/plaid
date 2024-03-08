package optest

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/resources"
	"github.com/stretchr/testify/require"
)

func (s *System) Observe(ctx context.Context, ref resources.Meta) *Observer {
	observer, created := s.observers.GetOrCreate(ref, func() *Observer {
		return NewObserver(s)
	})
	if created {
		token, err := s.observer.OnResource(ctx, ref, observer.onResourceEvent)
		require.NoError(s.t, err)
		observer.token = token
	}
	return observer
}

func (s *System) ObserveType(ctx context.Context, kind resources.Type) *Observer {
	observer, created := s.typeObservers.GetOrCreate(kind, func() *Observer {
		return NewObserver(s)
	})
	if created {
		token, err := s.observer.OnType(ctx, kind, observer.onResourceEvent)
		require.NoError(s.t, err)
		observer.token = token
	}
	return observer
}

type Observer struct {
	system   *System
	token    resources.WatchToken
	AnyEvent *ObserverAspect
	Create   *ObserverAspect
	Delete   *ObserverAspect
	Update   *ObserverAspect
	Status   *ObserverAspect
}

func NewObserver(s *System) *Observer {
	o := &Observer{
		system: s,
	}
	o.AnyEvent = &ObserverAspect{observer: o}
	o.Create = &ObserverAspect{observer: o}
	o.Delete = &ObserverAspect{observer: o}
	o.Update = &ObserverAspect{observer: o}
	o.Status = &ObserverAspect{observer: o}
	return o
}

func (o *Observer) onResourceEvent(ctx context.Context, event resources.ResourceChanged) error {
	o.AnyEvent.update()
	switch event.Operation {
	case resources.CreatedEvent:
		o.Create.update()
	case resources.DeletedEvent:
		o.Delete.update()
	case resources.UpdatedEvent:
		o.Update.update()
	case resources.StatusUpdated:
		o.Status.update()
	default:
		panic(fmt.Sprintf("Unknown event type %s", event.Operation))
	}
	return nil
}

func (o *Observer) consumeEvent(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-o.system.observer.Events():
		err := o.system.observer.Digest(ctx, e)
		return err
	}
}

type eventCounter uint64
