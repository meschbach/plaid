package optest

import (
	"context"
	"github.com/meschbach/plaid/internal/junk/jtest"
	"github.com/meschbach/plaid/resources"
	"github.com/stretchr/testify/require"
	"testing"
)

type System struct {
	t             *testing.T
	s             resources.System
	root          context.Context
	Legacy        *resources.TestSubsystem
	observer      resources.Watcher
	observers     *resources.MetaContainer[ObservedResource]
	typeObservers *resources.TypeContainer[ObservedType]
}

func (s *System) Observe(ctx context.Context, ref resources.Meta) *ObservedResource {
	observer, created := s.observers.GetOrCreate(ref, func() *ObservedResource {
		o := &ObservedResource{
			system: s,
		}
		o.AnyEvent = &ResourceAspect{observer: o}
		o.Status = &ResourceAspect{observer: o}
		return o
	})
	if created {
		token, err := s.observer.OnResource(ctx, ref, observer.onResourceEvent)
		require.NoError(s.t, err)
		observer.token = token
	}
	return observer
}

func (s *System) ObserveType(ctx context.Context, kind resources.Type) *ObservedType {
	observer, created := s.typeObservers.GetOrCreate(kind, func() *ObservedType {
		o := &ObservedType{
			system: s,
		}
		o.AnyEvent = &TypeAspect{observer: o}
		o.Create = &TypeAspect{observer: o}
		o.Delete = &TypeAspect{observer: o}
		return o
	})
	if created {
		token, err := s.observer.OnType(ctx, kind, observer.onResourceEvent)
		require.NoError(s.t, err)
		observer.token = token
	}
	return observer
}

func (s *System) MustCreate(ctx context.Context, ref resources.Meta, spec any) {
	storage, err := s.s.Storage(ctx)
	require.NoError(s.t, err)
	require.NoError(s.t, storage.Create(ctx, ref, spec))
}

func (s *System) MustDelete(ctx context.Context, ref resources.Meta) {
	exists, err := s.Legacy.Store.Delete(ctx, ref)
	require.NoError(s.t, err)
	require.True(s.t, exists, "must have existed")
}

func New(t *testing.T) (context.Context, *System) {
	ctx := jtest.ContextFromEnv(t)

	legacy := resources.WithTestSubsystem(t, ctx)
	systemObserver, err := legacy.Store.Watcher(ctx)
	require.NoError(t, err)
	sys := &System{
		t:             t,
		s:             legacy.System,
		root:          ctx,
		Legacy:        legacy,
		observer:      systemObserver,
		observers:     resources.NewMetaContainer[ObservedResource](),
		typeObservers: resources.NewTypeContainer[ObservedType](),
	}
	return ctx, sys
}

func From(t *testing.T, ctx context.Context, s resources.System) *System {
	storage, err := s.Storage(ctx)
	require.NoError(t, err)
	watcher, err := storage.Observer(ctx)
	require.NoError(t, err)

	return &System{
		t:             t,
		s:             s,
		root:          nil,
		Legacy:        nil,
		observer:      watcher,
		observers:     resources.NewMetaContainer[ObservedResource](),
		typeObservers: resources.NewTypeContainer[ObservedType](),
	}
}
