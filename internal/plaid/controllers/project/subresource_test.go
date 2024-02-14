package project

import (
	"context"
	"github.com/meschbach/plaid/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type ExampleSpec struct {
	Flag int `json:"flag"`
}

type ExampleStatus struct {
	Flag int `json:"flag"`
}

func TestClaimed(t *testing.T) {
	t.Run("Given a setup system", func(t *testing.T) {
		ctx, done := context.WithCancel(context.Background())
		t.Cleanup(done)

		claimed := resources.FakeMeta()
		core := resources.WithTestSubsystem(t, ctx)
		watcher, err := core.Store.Watcher(ctx)
		require.NoError(t, err)

		reconcile := 0
		env := &resourceEnv{
			which:   claimed,
			rpc:     core.Store,
			watcher: watcher,
			reconcile: func(ctx context.Context) error {
				reconcile++
				return nil
			},
		}
		t.Run("When a new claimed system is created", func(t *testing.T) {
			c := &Subresource[ExampleStatus]{}
			var status ExampleStatus
			step, err := c.Decide(ctx, env, &status)
			require.NoError(t, err)
			assert.Equal(t, SubresourceCreated, step, "Then thee result is create")

			t.Run("And the resource is created", func(t *testing.T) {
				ref := resources.FakeMeta()
				err := c.Create(ctx, env, ref, ExampleSpec{Flag: 10})
				require.NoError(t, err)

				t.Run("Then the resource is de-marked as created", func(t *testing.T) {
					step, err := c.Decide(ctx, env, &status)
					require.NoError(t, err)
					assert.Equal(t, SubresourceExists, step, "expected step exists, got %s", step)
				})
			})
		})
	})
}
