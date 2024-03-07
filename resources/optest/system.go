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
	storage       resources.Storage
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
		o.Update = &TypeAspect{observer: o}
		o.UpdateStatus = &TypeAspect{observer: o}
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
	require.NoError(s.t, s.storage.Create(ctx, ref, spec))
}

func (s *System) MustDelete(ctx context.Context, ref resources.Meta) {
	exists, err := s.storage.Delete(ctx, ref)
	require.NoError(s.t, err)
	require.True(s.t, exists, "must have existed")
}

func (s *System) MustUpdateStatus(ctx context.Context, ref resources.Meta, status interface{}) {
	exists, err := s.storage.UpdateStatus(ctx, ref, status)
	require.NoError(s.t, err)
	require.True(s.t, exists, "expected to exist but did not")
}

func (s *System) Run(name string, test func(t *testing.T, s *System)) {
	s.t.Run(name, func(t *testing.T) {
		next := &System{
			t:             t,
			s:             s.s,
			root:          s.root,
			Legacy:        s.Legacy,
			storage:       s.storage,
			observer:      s.observer,
			observers:     s.observers,
			typeObservers: s.typeObservers,
		}
		test(t, next)
	})
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
		storage:       legacy.Store,
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
		storage:       storage,
		observer:      watcher,
		observers:     resources.NewMetaContainer[ObservedResource](),
		typeObservers: resources.NewTypeContainer[ObservedType](),
	}
}
