package project

import (
	"context"
	"github.com/go-faker/faker/v4"
	"github.com/meschbach/plaid/internal/junk"
	"github.com/meschbach/plaid/internal/plaid/controllers/buildrun"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/internal/plaid/controllers/service"
	"github.com/meschbach/plaid/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestProjectAlpha1(t *testing.T) {
	t.Run("Given a Plaid instance with the configured controller", func(t *testing.T) {
		onDone := junk.SetupTestTracing(t)
		t.Cleanup(func() {
			shutdown, done := context.WithTimeout(context.Background(), 1*time.Second)
			defer done()

			onDone(shutdown)
		})

		baseCtx, done := context.WithTimeout(context.Background(), 1*time.Second)
		t.Cleanup(done)
		ctx, _ := junk.TraceSubtest(t, baseCtx, tracer)
		plaid := resources.WithTestSubsystem(t, ctx)
		plaid.AttachController("plaid.controllers.project", NewProjectSystem(plaid.Controller))

		t.Run("When a new project is created with a oneshot", func(t *testing.T) {
			ctx, _ = junk.TraceSubtest(t, ctx, tracer)
			tmpDir := os.TempDir()
			ctx, _ = junk.TraceSubtest(t, ctx, tracer)
			projectRef := resources.FakeMetaOf(Alpha1)
			projectSpec := Alpha1Spec{
				BaseDirectory: tmpDir,
				OneShots: []Alpha1OneShotSpec{
					{
						Name: "one-shot",
						Build: exec.TemplateAlpha1Spec{
							Command: "builder",
						},
						Run: exec.TemplateAlpha1Spec{
							Command: "run",
						},
					},
				},
			}
			require.NoError(t, plaid.Store.Create(ctx, projectRef, projectSpec))

			t.Run("Then a buildrun is created", func(t *testing.T) {
				ctx, _ = junk.TraceSubtest(t, ctx, tracer)
				var projectStatus Alpha1Status
				success, err := resources.WaitOn(ctx, resources.ForStatusState(plaid.Store, projectRef, &projectStatus, func(status *Alpha1Status) (bool, error) {
					if len(status.OneShots) < 1 {
						return false, nil
					}
					return status.OneShots[0].Ref != nil, nil
				}))
				require.NoError(t, err)
				assert.True(t, success, "successfully retrieved project status")

				var spec buildrun.AlphaSpec1
				exists, problem := plaid.Store.Get(ctx, *projectStatus.OneShots[0].Ref, &spec)
				require.NoError(t, problem)
				require.True(t, exists, "resource must still exist")

				t.Run("with the correct build configuration", func(t *testing.T) {
					assert.Equal(t, tmpDir, spec.Build.WorkingDir, "working directory set correctly")
					assert.Equal(t, "builder", spec.Build.Command, "builder command set correctly")
				})
				t.Run("with the correct run configuration", func(t *testing.T) {
					assert.Equal(t, tmpDir, spec.Run.WorkingDir, "working directory set correctly")
					assert.Equal(t, "run", spec.Run.Command, "run command set correctly")
				})
			})

			t.Run("And the buildrun exits for both build and run", func(t *testing.T) {
				ctx, _ = junk.TraceSubtest(t, ctx, tracer)
				var buildRunRef *resources.Meta
				var buildRunStatus Alpha1Status
				success, err := resources.WaitOn(ctx, resources.ForStatusState(plaid.Store, projectRef, &buildRunStatus, func(status *Alpha1Status) (bool, error) {
					buildRunRef = status.OneShots[0].Ref
					return buildRunRef != nil, nil
				}))
				require.NoError(t, err)
				require.True(t, success)
				br := &BuildRunStatusMock{
					Ref:   buildRunRef,
					Store: plaid.Store,
				}
				require.NoError(t, br.FinishNow(ctx, 0))

				t.Run("Then the exit is noted in the project status", func(t *testing.T) {
					ctx, _ = junk.TraceSubtest(t, ctx, tracer)
					var status Alpha1Status
					exists, err := resources.WaitOn(ctx, resources.ForStatusState[Alpha1Status](plaid.Store, projectRef, &status, func(status *Alpha1Status) (bool, error) {
						return status.Done, nil
					}))
					require.NoError(t, err, "Error with status %#v", status)
					require.True(t, exists, "must exist")

					if assert.Len(t, status.OneShots, 1) {
						assert.True(t, status.OneShots[0].Done, "must have completed")
					}

					assert.True(t, status.Done, "project is completed.")
				})
			})
		})

		t.Run("When a new project is created with a daemon service", func(t *testing.T) {
			tmpDir := os.TempDir()
			ctx, _ = junk.TraceSubtest(t, ctx, tracer)
			projectRef := resources.FakeMetaOf(Alpha1)
			builderCommand := faker.Word()
			runCommand := faker.Word()
			projectSpec := Alpha1Spec{
				BaseDirectory: tmpDir,
				Daemons: []Alpha1DaemonSpec{
					{
						Name: "daemon",
						Build: &exec.TemplateAlpha1Spec{
							Command: builderCommand,
						},
						Run: exec.TemplateAlpha1Spec{
							Command: runCommand,
						},
					},
				},
			}
			require.NoError(t, plaid.Store.Create(ctx, projectRef, projectSpec))

			t.Run("Then a buildrun is created", func(t *testing.T) {
				ctx, _ = junk.TraceSubtest(t, ctx, tracer)
				var found []resources.Meta
				for len(found) == 0 {
					var err error
					found, err = plaid.Store.FindClaimedBy(ctx, projectRef, []resources.Type{service.Alpha1})
					require.NoError(t, err)

					time.Sleep(25 * time.Millisecond)
				}
				if assert.Len(t, found, 1) {
					var spec buildrun.AlphaSpec1
					exists, problem := plaid.Store.Get(ctx, found[0], &spec)
					require.NoError(t, problem)
					require.True(t, exists, "resource must still exist")

					t.Run("with the correct build configuration", func(t *testing.T) {
						assert.Equal(t, tmpDir, spec.Build.WorkingDir, "working directory set correctly")
						assert.Equal(t, builderCommand, spec.Build.Command, "builder command set correctly")
					})
					t.Run("with the correct run configuration", func(t *testing.T) {
						assert.Equal(t, tmpDir, spec.Run.WorkingDir, "working directory set correctly")
						assert.Equal(t, runCommand, spec.Run.Command, "run command set correctly")
					})
				}
			})
		})
	})
}

func RequireGetStatus[T any](t *testing.T, ctx context.Context, store *resources.Client, ref resources.Meta) *T {
	var status T
	exists, err := store.GetStatus(ctx, ref, &status)
	require.NoError(t, err)
	require.True(t, exists, "must exists")
	return &status
}
