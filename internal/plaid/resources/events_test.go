package resources

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestEventsSystem(t *testing.T) {
	ctx, done := context.WithTimeout(context.Background(), 1*time.Second)
	defer done()

	res := WithTestSubsystem(t, ctx)
	t.Cleanup(res.SystemDone)

	t.Run("Given an object", func(t *testing.T) {
		ref := FakeMeta()
		require.NoError(t, res.Store.Create(ctx, ref, ExampleResource{Enabled: true}))

		t.Run("When an info event is recorded against it", func(t *testing.T) {
			exists, err := res.Store.Log(ctx, ref, Info, "some data %s", "arg")
			require.NoError(t, err)
			require.True(t, exists, "should not have been deleted.")

			t.Run("Then the event get be retrieved with All", func(t *testing.T) {
				events, exists, err := res.Store.GetLogs(ctx, ref, AllEvents)
				require.NoError(t, err)
				require.True(t, exists, "must exist")
				if assert.Len(t, events, 1) {

				}
			})
		})
	})
}

type ExampleResource struct {
	Enabled bool
}
