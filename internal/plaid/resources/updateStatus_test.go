package resources

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestUpdateStatus(t *testing.T) {
	ctx, done := context.WithTimeout(context.Background(), 1*time.Second)
	defer done()

	res := WithTestSubsystem(t, ctx)
	t.Cleanup(res.SystemDone)

	exampleMeta := Meta{
		Type: Type{
			Kind:    "example.plaid.meschbach.com",
			Version: "Alpha1",
		},
		Name: "some-random-name",
	}

	r := res.Controller.Client()
	require.NoError(t, r.Create(ctx, exampleMeta, updateSpec{}))
	exists, err := r.UpdateStatus(ctx, exampleMeta, updateStatus{Ready: true})
	require.NoError(t, err)
	require.True(t, exists, "resource should exist")
}

type updateSpec struct {
	Enabled bool `json:"enabled"`
}

type updateStatus struct {
	Ready bool `json:"ready"`
}
