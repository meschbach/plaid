package project

import (
	"context"
	"github.com/go-faker/faker/v4"
	"github.com/meschbach/plaid/controllers/tooling"
	"github.com/meschbach/plaid/internal/plaid/controllers/service"
	"github.com/meschbach/plaid/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDaemonState(t *testing.T) {
	t.Run("Given a new project space", func(t *testing.T) {
		testCtx, done := context.WithCancel(context.Background())
		t.Cleanup(done)

		plaid := resources.WithTestSubsystem(t, testCtx)
		watcher, err := plaid.Store.Watcher(testCtx)
		require.NoError(t, err)

		reconciled := 0
		exampleProject := resources.Meta{Type: Alpha1, Name: faker.Name()}
		env := tooling.Env{
			Subject: exampleProject,
			Storage: plaid.Store,
			Watcher: watcher,
			Reconcile: func(ctx context.Context) error {
				reconciled++
				return nil
			},
		}

		daemon := &daemonState{}

		exampleDaemonSpec := Alpha1DaemonSpec{}
		exampleSpec := Alpha1Spec{}

		t.Run("When asked for next steps", func(t *testing.T) {
			step, err := daemon.decideNextStep(testCtx, env, "")
			require.NoError(t, err)
			assert.Equal(t, daemonCreate, step, "Then we will create our resources")
		})

		t.Run("When Created", func(t *testing.T) {
			err := daemon.create(testCtx, env, exampleSpec, exampleDaemonSpec)
			require.NoError(t, err)
			step, err := daemon.decideNextStep(testCtx, env, "")
			require.NoError(t, err)

			assert.Equal(t, daemonWait, step, "Then the next step is wait, got %s", step)
			status := &Alpha1DaemonStatus{}
			daemon.toStatus(exampleDaemonSpec, status)
			assert.Equal(t, daemon.service.Ref, *status.Current, "Then reports correct service Ref")
			assert.False(t, status.Ready, "Then it is not ready")
		})

		t.Run("When daemon has become ready", func(t *testing.T) {
			exists, err := plaid.Store.UpdateStatus(testCtx, daemon.service.Ref, service.Alpha1Status{
				Ready: true,
			})
			require.NoError(t, err, "failed to update status to ready")
			require.True(t, exists, "must exist")

			step, err := daemon.decideNextStep(testCtx, env, "")
			require.NoError(t, err)

			assert.Equal(t, daemonWait, step, "Then the next step is to wait")
			status := &Alpha1DaemonStatus{}
			daemon.toStatus(exampleDaemonSpec, status)
			assert.Equal(t, daemon.service.Ref, *status.Current, "Then reports correct service Ref")
			assert.True(t, status.Ready, "Then it is ready")
		})
	})
}
