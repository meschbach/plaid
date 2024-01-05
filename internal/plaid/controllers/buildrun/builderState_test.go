package buildrun

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/internal/plaid/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestBuilderState(t *testing.T) {
	t.Run("Given a new resource system and a builder", func(t *testing.T) {
		timedContext, onTimerDone := context.WithTimeout(context.Background(), 1*time.Second)
		t.Cleanup(onTimerDone)
		worldState := resources.WithTestSubsystem(t, timedContext)
		subject := builderState{}

		t.Run("When the builder is asked for the next step", func(t *testing.T) {
			nextStep, _, err := subject.decideNextStep(timedContext, worldState.Store)
			require.NoError(t, err)

			t.Run("Then creation is the next step recommended", func(t *testing.T) {
				assert.Equal(t, builderNextCreate, nextStep)
			})
		})
		t.Run("When the builder is instructed to create a new resource", func(t *testing.T) {
			err := subject.create(timedContext, &stateEnv{
				object: resources.Meta{
					Type: Alpha1,
					Name: "subject-under-test",
				},
				rpc: worldState.Store,
			}, exec.TemplateAlpha1Spec{
				Command:    "test",
				WorkingDir: "/non-existent",
			})
			require.NoError(t, err)

			t.Run("Then a new invocation resources has been created", func(t *testing.T) {
				var proc exec.InvocationAlphaV1Spec
				exists, err := worldState.Store.Get(timedContext, subject.lastBuild, &proc)
				require.NoError(t, err)
				if assert.True(t, exists, "resource exists") {
					assert.Equal(t, "test", proc.Exec)
				}
			})

			t.Run("Then the builder recommends waiting", func(t *testing.T) {
				nextStep, _, err := subject.decideNextStep(timedContext, worldState.Store)
				require.NoError(t, err)

				assert.Equal(t, builderNextWait, nextStep)
			})

			t.Run("And the program finishes with exit status 0", func(t *testing.T) {
				started := time.Now()
				exitStatus := 0
				exists, err := worldState.Store.UpdateStatus(timedContext, subject.lastBuild, exec.InvocationAlphaV1Status{
					Started:    &started,
					Finished:   &started,
					ExitStatus: &exitStatus,
				})
				require.NoError(t, err)
				require.True(t, exists)

				t.Run("Then the builder states it has completed successfully", func(t *testing.T) {
					next, _, err := subject.decideNextStep(timedContext, worldState.Store)
					require.NoError(t, err)
					assert.Equal(t, builderStateSuccessfullyCompleted, next, "Unexpected recommendation: %s", next)
				})
			})
		})
	})
}
