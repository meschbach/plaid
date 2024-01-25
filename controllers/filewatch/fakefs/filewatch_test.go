package fakefs

import (
	"context"
	"github.com/meschbach/plaid/controllers/filewatch"
	"github.com/meschbach/plaid/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestFileWatch(t *testing.T) {
	t.Run("Given a filewatch controller backed by the fakefs", func(t *testing.T) {
		ctx, done := context.WithCancel(context.Background())
		t.Cleanup(done)

		res := resources.WithTestSubsystem(t, ctx)
		watcher, err := res.Store.Watcher(ctx)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, watcher.Close(ctx))
		})

		fake := New()

		res.AttachController("fakefs", fake)

		watchResources := filewatch.NewController(res.Controller, fake)
		res.AttachController("filewatch", watchResources)

		t.Run("When a new path is watched", func(t *testing.T) {
			examplePath := "/tmp/example"

			meta := resources.FakeMetaOf(filewatch.Alpha1)
			var lastStatus filewatch.Alpha1Status
			lastUpdated := time.Now()
			_, err := watcher.OnResourceStatusChanged(ctx, meta, func(ctx context.Context, changed resources.ResourceChanged) error {
				lastUpdated = time.Now()
				_, err := res.Store.GetStatus(ctx, meta, &lastStatus)
				return err
			})
			require.NoError(t, err)

			created := time.Now()
			require.NoError(t, res.Store.Create(ctx, meta, filewatch.Alpha1Spec{
				AbsolutePath: examplePath,
			}))
			//wait for status update
			for lastUpdated.Before(created) {
				select {
				case <-ctx.Done():
					require.NoError(t, ctx.Err())
				case e := <-watcher.Feed:
					require.NoError(t, watcher.Digest(ctx, e))
				}
			}

			t.Run("And the root path changes", func(t *testing.T) {
				updated := time.Now()
				require.NoError(t, fake.FileModified(ctx, examplePath, updated))

				//wait for status update
				for lastUpdated.Before(updated) {
					select {
					case <-ctx.Done():
						require.NoError(t, ctx.Err())
					case e := <-watcher.Feed:
						require.NoError(t, watcher.Digest(ctx, e))
					}
				}

				assert.True(t, updated.Equal(*lastStatus.LastChange), "Then the last changed time is equal to the updated time")
			})

			t.Run("And a sub-path changes", func(t *testing.T) {
				updated := time.Now()
				require.NoError(t, fake.FileModified(ctx, examplePath+"/subpath/file", updated))

				//wait for status update
				for lastUpdated.Before(updated) {
					select {
					case <-ctx.Done():
						require.NoError(t, ctx.Err())
					case e := <-watcher.Feed:
						require.NoError(t, watcher.Digest(ctx, e))
					}
				}

				assert.True(t, updated.Equal(*lastStatus.LastChange), "Then the last changed time is equal to the updated time")
			})
		})
	})
}
