package fakefs

import (
	"github.com/meschbach/plaid/controllers/filewatch"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/optest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestFileWatch(t *testing.T) {
	t.Run("Given a filewatch controller backed by the fakefs", func(t *testing.T) {
		ctx, sys := optest.New(t)
		fake := New()

		sys.Legacy.AttachController("fakefs", fake)

		watchResources := filewatch.NewController(sys.Legacy.System, fake)
		sys.Legacy.AttachController("filewatch", watchResources)

		t.Run("When a new path is watched", func(t *testing.T) {
			examplePath := "/tmp/example"

			meta := resources.FakeMetaOf(filewatch.Alpha1)
			target := sys.Observe(ctx, meta)
			statusChange := target.Status.Fork()

			sys.MustCreate(ctx, meta, filewatch.Alpha1Spec{
				AbsolutePath: examplePath,
			})

			statusChange.Wait(t, ctx)

			t.Run("And the root path changes", func(t *testing.T) {
				updated := time.Now()
				fsPush := target.Status.Fork()
				require.NoError(t, fake.FileModified(ctx, examplePath, updated))

				fsPush.Wait(t, ctx)

				result := optest.MustGetStatus[filewatch.Alpha1Status](sys, meta)
				if assert.NotNil(t, result.LastChange) {
					assert.True(t, result.LastChange.Equal(updated) || result.LastChange.After(updated), "must be equal or after the updated time")
				}
			})

			t.Run("And a sub-path changes", func(t *testing.T) {
				fsSubpathPush := target.Status.Fork()
				updated := time.Now()
				require.NoError(t, fake.FileModified(ctx, examplePath+"/subpath/file", updated))

				fsSubpathPush.Wait(t, ctx)
				result := optest.MustGetStatus[filewatch.Alpha1Status](sys, meta)
				if assert.NotNil(t, result.LastChange) {
					assert.True(t, result.LastChange.Equal(updated) || result.LastChange.After(updated), "must be equal or after the updated time")
				}
			})
		})
	})
}
