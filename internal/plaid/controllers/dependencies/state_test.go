package dependencies

import (
	"context"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"github.com/meschbach/go-junk-bucket/testing/faking"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestState(t *testing.T) {
	t.Run("Given an empty dependency set", func(t *testing.T) {
		ctx, testDone := context.WithCancel(context.Background())
		t.Cleanup(testDone)

		world := resources.WithTestSubsystem(t, ctx)
		watcher, err := world.Store.Watcher(ctx)
		require.NoError(t, err)

		reconciledCalledCount := 0
		env := Env{
			Storage: world.Store,
			Watcher: watcher,
			OnChange: func(ctx context.Context) error {
				reconciledCalledCount++
				return nil
			},
		}
		s := State{}
		s.Init(nil)

		t.Run("When asked to reconcile", func(t *testing.T) {
			ready, status, err := s.Reconcile(ctx, env)
			require.NoError(t, err)
			assert.True(t, ready, "Then is ready")
			assert.Empty(t, status, "then no required is provided")
		})
	})

	t.Run("Given a dependency set before creation", func(t *testing.T) {
		ctx, testDone := context.WithCancel(context.Background())
		t.Cleanup(testDone)

		world := resources.WithTestSubsystem(t, ctx)
		watcher, err := world.Store.Watcher(ctx)
		require.NoError(t, err)

		reconciledCalledCount := 0
		env := Env{
			Storage: world.Store,
			Watcher: watcher,
			OnChange: func(ctx context.Context) error {
				reconciledCalledCount++
				return nil
			},
		}
		s := State{}

		fakedType := resources.FakeType()
		words := faking.DistinctWords(3)
		firePit := resources.Meta{Type: fakedType, Name: words[0]}
		firewood := resources.Meta{Type: fakedType, Name: words[1]}
		spark := resources.Meta{Type: fakedType, Name: words[2]}

		s.Init([]NamedDependencyAlpha1{
			{firePit.Name, firePit},
			{firewood.Name, firewood},
			{spark.Name, spark},
		})

		t.Run("When asked to reconcile", func(t *testing.T) {
			ready, status, err := s.Reconcile(ctx, env)
			require.NoError(t, err)
			assert.False(t, ready, "Then is not ready")
			assert.Len(t, status, 3, "then provides all required")
			assert.Equal(t, 3, reconciledCalledCount, "Then asked to requeue reconciliation")

			t.Run("And asked to reconcile again", func(t *testing.T) {
				ready, status, err := s.Reconcile(ctx, env)
				require.NoError(t, err)
				assert.False(t, ready, "Then is not ready")
				assert.Len(t, status, 3, "then provides all required")
				assert.Equal(t, 3, reconciledCalledCount, "Then not asked to reconcile yet again")
			})
		})
	})
}
