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
	observers     *resources.MetaContainer[Observer]
	typeObservers *resources.TypeContainer[Observer]
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

func (s *System) Run(name string, test func(t *testing.T, s *System, ctx context.Context)) {
	s.t.Run(name, func(t *testing.T) {
		ctx, done := context.WithCancel(s.root)
		t.Cleanup(done)

		next := &System{
			t:             t,
			s:             s.s,
			root:          ctx,
			Legacy:        s.Legacy,
			storage:       s.storage,
			observer:      s.observer,
			observers:     s.observers,
			typeObservers: s.typeObservers,
		}
		test(t, next, ctx)
	})
}

func New(t *testing.T) (context.Context, *System) {
	envCtx := jtest.ContextFromEnv(t)
	ctx, done := context.WithCancel(envCtx)
	t.Cleanup(done)

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
		observers:     resources.NewMetaContainer[Observer](),
		typeObservers: resources.NewTypeContainer[Observer](),
	}
	return ctx, sys
}

func From(t *testing.T, parent context.Context, s resources.System) *System {
	ctx, done := context.WithCancel(parent)
	t.Cleanup(done)

	storage, err := s.Storage(ctx)
	require.NoError(t, err)
	watcher, err := storage.Observer(ctx)
	require.NoError(t, err)

	return &System{
		t:             t,
		s:             s,
		root:          ctx,
		Legacy:        nil,
		storage:       storage,
		observer:      watcher,
		observers:     resources.NewMetaContainer[Observer](),
		typeObservers: resources.NewTypeContainer[Observer](),
	}
}
