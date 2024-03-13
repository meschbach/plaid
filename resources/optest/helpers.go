package optest

import (
	"context"
	"github.com/meschbach/plaid/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func MustGetStatus[Status any](s *System, ref resources.Meta) Status {
	var out Status
	exists, err := s.storage.GetStatus(s.root, ref, &out)
	require.NoError(s.t, err, "failed to retrieve status of %s", ref)
	require.True(s.t, exists, "resource %s was expected to exist and status retrieval but was not", ref)
	return out
}

func MustGetSpec[Spec any](s *System, ref resources.Meta) Spec {
	var out Spec
	exists, err := s.storage.Get(s.root, ref, &out)
	require.NoError(s.t, err, "failed to retrieve status of %s", ref)
	require.True(s.t, exists, "resource %s was expected to exist and status retrieval but was not", ref)
	return out
}

func MustUpdateStatusAndWait[Status any](s *System, change *ObserverAspect, update resources.Meta, status Status) {
	point := change.Fork()
	MustUpdateStatus[Status](s, update, status)
	point.Wait(s.t, s.root)
}

func MustUpdateStatusRaw(t *testing.T, ctx context.Context, store *resources.Client, ref resources.Meta, status any) {
	exists, err := store.UpdateStatus(ctx, ref, status)
	require.NoError(t, err, "update must not error")
	require.True(t, exists, "build must exist")
}

func MustUpdateStatus[Status any](s *System, update resources.Meta, status Status) {
	MustUpdateStatusRaw(s.t, s.root, s.Legacy.Store, update, status)
}

func (s *System) MustUpdateAndWait(changePoint *ObserverAspect, which resources.Meta, newSpec interface{}) {
	s.t.Helper()
	changeWatcher := changePoint.Fork()
	exists, err := s.storage.Update(s.root, which, newSpec)
	require.NoError(s.t, err)
	require.True(s.t, exists)
	changeWatcher.Wait(s.t, s.root)
}

func MustBeMissingSpec[Spec any](plaid *System, ref resources.Meta, messageAndValues ...any) bool {
	var spec Spec
	exists, err := plaid.storage.Get(plaid.root, ref, &spec)
	if !assert.NoError(plaid.t, err) {
		return false
	}
	return assert.False(plaid.t, exists, messageAndValues)
}
