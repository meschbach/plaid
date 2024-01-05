package projectfile

import (
	"context"
	"github.com/go-faker/faker/v4"
	"github.com/meschbach/plaid/internal/plaid/controllers/project"
	"github.com/meschbach/plaid/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestProjectStateDecideNext(t *testing.T) {
	t.Run("Given an new state", func(t *testing.T) {
		testContext, testDone := context.WithCancel(context.Background())
		t.Cleanup(testDone)

		plaid := resources.WithTestSubsystem(t, testContext)
		watcher, err := plaid.Store.Watcher(testContext)
		require.NoError(t, err)

		env := stateEnv{
			which:   resources.FakeMeta(),
			rpc:     plaid.Store,
			watcher: watcher,
			reconcile: func(ctx context.Context) error {
				return nil
			},
		}
		p := &projectState{}

		t.Run("Then the next state ", func(t *testing.T) {
			next, err := p.decideNextSteps(testContext, env)
			require.NoError(t, err)
			assert.Equal(t, projectNextCreate, next, "initial state should be next")
		})

		t.Run("When the project spec has a single one shot", func(t *testing.T) {
			spec := Alpha1Spec{
				WorkingDirectory: "/tmp/example",
				ProjectFile:      "plaid.json",
			}
			truePointer := true
			configFile := Configuration{
				OneShot: &truePointer,
				Name:    faker.Word(),
				Run:     "echo test",
			}
			require.NoError(t, p.create(testContext, env, spec, configFile))

			t.Run("Then the next step is to wait", func(t *testing.T) {
				next, err := p.decideNextSteps(testContext, env)
				require.NoError(t, err)
				assert.Equal(t, projectNextWait, next, "expected wait, got %s", next)
			})

			t.Run("Then a build run is listed as the current", func(t *testing.T) {
				var status Alpha1Status
				p.updateStatus(testContext, &status)

				assert.NotNil(t, status.Current, "should have a current ref for the build run")
			})
		})

		t.Run("When the underlying project completes successfully", func(t *testing.T) {
			var status Alpha1Status
			p.updateStatus(testContext, &status)

			projectStatus := project.Alpha1Status{
				Done:   true,
				Result: project.Alpha1StateSuccess,
			}
			exists, err := plaid.Store.UpdateStatus(testContext, *status.Current, projectStatus)
			require.NoError(t, err)
			require.True(t, exists, "resource %s must exist", *status.Current)

			t.Run("Then the project is complete", func(t *testing.T) {
				step, err := p.decideNextSteps(testContext, env)
				require.NoError(t, err)
				assert.Equal(t, projectNextDone, step, "expected done, got %s", step)
			})

			t.Run("Then the ProjectFile status is done with success", func(t *testing.T) {
				p.updateStatus(testContext, &status)

				assert.True(t, status.Done, "project is done")
				assert.True(t, status.Success, "project should be completed successfully: %+v", p)
			})
		})
	})
}

func intPointer(i int) *int {
	return &i
}
func nowPointer() *time.Time {
	t := time.Now()
	return &t
}
