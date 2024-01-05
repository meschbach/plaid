package resources

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thejerf/suture/v4"
	"testing"
)

func TestWatcherSubsystem(t *testing.T) {
	t.Run("Given a watcher", func(t *testing.T) {
		ctx, done := context.WithCancel(context.Background())
		t.Cleanup(done)

		core := NewController()
		client := core.Client()

		s := suture.NewSimple("core")
		s.Add(core)
		s.ServeBackground(ctx)

		t.Run("When watching all events", func(t *testing.T) {
			subject, err := client.Watcher(ctx)
			require.NoError(t, err)

			consumerCount := 0
			_, err = subject.OnAll(ctx, func(ctx context.Context, changed ResourceChanged) error {
				consumerCount++
				return nil
			})
			require.NoError(t, err)

			meta := FakeMeta()
			require.NoError(t, client.Create(ctx, meta, ExampleResource{Enabled: false}))

			require.NoError(t, digestAllPending(ctx, subject))

			assert.Equal(t, 1, consumerCount, "Then the consumer is invoked")
		})

		t.Run("When listening for status updates on a resource", func(t *testing.T) {
			consumedCount := 0
			watchingRef := FakeMeta()

			w, err := client.Watcher(ctx)
			require.NoError(t, err)
			_, err = w.OnResourceStatusChanged(ctx, watchingRef, func(ctx context.Context, changed ResourceChanged) error {
				consumedCount++
				return nil
			})
			require.NoError(t, err)

			t.Run("And a change to an unrelated resource of the same resource occurs", func(t *testing.T) {
				beforeCount := consumedCount

				other := FakeMetaOf(watchingRef.Type)
				require.NoError(t, client.Create(ctx, other, ExampleResource{Enabled: false}))
				require.NoError(t, digestAllPending(ctx, w))

				assert.Equal(t, beforeCount, consumedCount, "then the consumer is not notified")
			})

			t.Run("And the resource gets created", func(t *testing.T) {
				beforeCount := consumedCount
				require.NoError(t, client.Create(ctx, watchingRef, ExampleResource{Enabled: true}))
				require.NoError(t, digestAllPending(ctx, w))

				assert.Equal(t, beforeCount, consumedCount, "Then the consumer is not notified")
			})

			t.Run("And a change to the status occurs", func(t *testing.T) {
				beforeCount := consumedCount
				exists, err := client.UpdateStatus(ctx, watchingRef, ExampleResource{Enabled: true})
				require.NoError(t, err)
				require.True(t, exists)
				require.NoError(t, digestAllPending(ctx, w))

				assert.Less(t, beforeCount, consumedCount, "Then the consumer is notified")
			})

			t.Run("And the resource is deleted", func(t *testing.T) {
				beforeCount := consumedCount
				exists, err := client.Delete(ctx, watchingRef)
				require.NoError(t, err)
				require.True(t, exists)
				require.NoError(t, digestAllPending(ctx, w))

				assert.Less(t, beforeCount, consumedCount, "Then the consumer is notified")
			})
		})
	})
}

func digestAllPending(ctx context.Context, subject *ClientWatcher) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-subject.Feed:
			if err := subject.Digest(ctx, e); err != nil {
				return err
			}
		default:
			return nil
		}
	}
}
