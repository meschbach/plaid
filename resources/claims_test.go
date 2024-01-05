package resources

import (
	"context"
	"github.com/meschbach/plaid/internal/junk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestResourceOwnership(t *testing.T) {
	done := junk.SetupTestTracing(t)
	defer func() {
		shutdown, complete := context.WithTimeout(context.Background(), 1*time.Second)
		defer complete()

		done(shutdown)
	}()
	testCtx, closeTestContext := context.WithCancel(context.Background())
	t.Cleanup(closeTestContext)
	plaid := WithTestSubsystem(t, testCtx)
	client := plaid.Controller.Client()

	t.Run("Given an existing resource", func(t *testing.T) {
		ctx, _ := junk.TraceSubtest(t, testCtx, tracing)
		existingRef := FakeMeta()
		require.NoError(t, client.Create(ctx, existingRef, existingRef))

		watcher, err := plaid.Store.Watcher(ctx)
		require.NoError(t, err)
		_, err = watcher.OnResource(ctx, existingRef, func(ctx context.Context, changed ResourceChanged) error {
			switch changed.Operation {
			default:
				return nil
			}
		})
		require.NoError(t, err)

		t.Run("When claimed by another resource", func(t *testing.T) {
			ctx, _ = junk.TraceSubtest(t, ctx, tracing)
			//todo: test case for nonexistent resource
			claimer := FakeMeta()
			require.NoError(t, client.Create(ctx, claimer, claimer))
			exists, err := client.Claims(ctx, existingRef, claimer)
			require.NoError(t, err)
			assert.True(t, exists, "exists")

			t.Run("Then the resource is locatable by the resource", func(t *testing.T) {
				ctx, _ = junk.TraceSubtest(t, ctx, tracing)

				found, err := client.FindClaimedBy(ctx, claimer, nil)
				require.NoError(t, err)
				assert.Len(t, found, 1)
			})
		})
	})
}
