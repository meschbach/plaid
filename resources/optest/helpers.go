package optest

import (
	"context"
	"github.com/meschbach/plaid/resources"
	"github.com/stretchr/testify/require"
	"testing"
)

func MustGetStatus[Status any](s *System, ref resources.Meta) Status {
	var out Status
	exists, err := s.Legacy.Store.GetStatus(s.root, ref, &out)
	require.NoError(s.t, err, "failed to retrieve status of %s", ref)
	require.True(s.t, exists, "resource %s was expected to exist and status retrieval but was not", ref)
	return out
}

func MustGetSpec[Spec any](s *System, ref resources.Meta) Spec {
	var out Spec
	exists, err := s.Legacy.Store.Get(s.root, ref, &out)
	require.NoError(s.t, err, "failed to retrieve status of %s", ref)
	require.True(s.t, exists, "resource %s was expected to exist and status retrieval but was not", ref)
	return out
}

func MustUpdateStatusAndWait[Status any](s *System, change *ResourceAspect, update resources.Meta, status Status) {
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

func (s *System) MustUpdateAndWait(changePoint *ResourceAspect, which resources.Meta, newSpec interface{}) {
	changeWatcher := changePoint.Fork()
	exists, err := s.storage.Update(s.root, which, newSpec)
	require.NoError(s.t, err)
	require.True(s.t, exists)
	changeWatcher.Wait(s.t, s.root)
}
