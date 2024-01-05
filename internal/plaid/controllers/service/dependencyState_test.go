package service

import (
	"context"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources/operator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type exampleResource struct {
	Increment int `json:"increment"`
}

func TestDependencyState(t *testing.T) {
	t.Run("Given a system with a dependency on a missing resource", func(t *testing.T) {
		timedContext, timedContextDone := context.WithTimeout(context.Background(), 1*time.Second)
		t.Cleanup(timedContextDone)

		world := resources.WithTestSubsystem(t, timedContext)
		observer := resources.FakeMeta()
		target := resources.FakeMeta()
		dep := &dependencyState{
			ref: target,
		}
		watcher, err := world.Store.Watcher(timedContext)
		require.NoError(t, err)

		reconciledCalledCount := 0
		env := resEnv{
			object:  observer,
			rpc:     world.Store,
			watcher: watcher,
			reconcile: func(ctx context.Context) error {
				reconciledCalledCount++
				return nil
			},
		}

		t.Run("When asked to decide the next step", func(t *testing.T) {
			step, err := dep.decideNextStep(timedContext, env)
			require.NoError(t, err)

			t.Run("Then it is to create the watch", func(t *testing.T) {
				assert.Equal(t, dependencySetup, step)
			})

			t.Run("And setup", func(t *testing.T) {
				require.NoError(t, dep.setup(timedContext, env))

				t.Run("Then the next step is to wait", func(t *testing.T) {
					step, err := dep.decideNextStep(timedContext, env)
					require.NoError(t, err)
					assert.Equal(t, dependencyWait, step)
				})
			})
		})

		t.Run("When the target resource status is created as ready", func(t *testing.T) {
			require.NoError(t, world.Store.Create(timedContext, target, exampleResource{Increment: 0}))
			e, err := world.Store.UpdateStatus(timedContext, target, operator.ReadyStatus{Ready: true})
			require.NoError(t, err)
			require.True(t, e, "update status on target exists")

			t.Run("And the watch events are consumed", func(t *testing.T) {
				oldReconcileCount := reconciledCalledCount
				for oldReconcileCount == reconciledCalledCount {
					select {
					case e := <-watcher.Feed:
						require.NoError(t, watcher.Digest(timedContext, e), "dispatching watching events")
					case <-timedContext.Done():
						panic(timedContext.Err())
					}
				}

				t.Run("Then the reconciliation function is invoked", func(t *testing.T) {
					assert.Equal(t, 1, reconciledCalledCount, "reconciled called")
				})

				t.Run("Then the resource is ready", func(t *testing.T) {
					step, err := dep.decideNextStep(timedContext, env)
					require.NoError(t, err)
					assert.Equal(t, dependencyReady, step, "should be ready but was %s", step)
				})
			})
		})
	})
}
