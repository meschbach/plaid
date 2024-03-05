package optest

import (
	"context"
	"github.com/meschbach/plaid/internal/junk/jtest"
	"github.com/meschbach/plaid/resources"
	"github.com/stretchr/testify/require"
	"testing"
)

type System struct {
	t         *testing.T
	root      context.Context
	Legacy    *resources.TestSubsystem
	observer  *resources.ClientWatcher
	observers *resources.MetaContainer[ObservedResource]
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

func (s *System) MustCreate(ctx context.Context, ref resources.Meta, spec any) {
	require.NoError(s.t, s.Legacy.Store.Create(ctx, ref, spec))
}

func New(t *testing.T) (context.Context, *System) {
	ctx := jtest.ContextFromEnv(t)

	legacy := resources.WithTestSubsystem(t, ctx)
	systemObserver, err := legacy.Store.Watcher(ctx)
	require.NoError(t, err)
	sys := &System{
		t:         t,
		root:      ctx,
		Legacy:    legacy,
		observer:  systemObserver,
		observers: resources.NewMetaContainer[ObservedResource](),
	}
	return ctx, sys
}
