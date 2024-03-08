package project

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/controllers/buildrun"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/optest"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestProjectAlpha1OneShot(t *testing.T) {
	t.Run("Given a Plaid instance with the configured controller", func(t *testing.T) {
		ctx, plaid := optest.New(t)
		plaid.Legacy.AttachController("plaid.controllers.project", NewProjectSystem(plaid.Legacy.Controller))

		t.Run("When a new project is Created with a oneshot", func(t *testing.T) {
			tmpDir := os.TempDir()
			projectRef := resources.FakeMetaOf(Alpha1)
			projectWatcher := plaid.Observe(ctx, projectRef)

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
			creationStatus := projectWatcher.Status.Fork()
			plaid.MustCreate(ctx, projectRef, projectSpec)

			t.Run("Then a buildrun is Created", func(t *testing.T) {
				creationStatus.Wait(t, ctx)
				projectStatus := optest.MustGetStatus[Alpha1Status](plaid, projectRef)
				maybeOneShotRef := projectStatus.OneShots[0].Ref
				if !assert.NotNil(t, maybeOneShotRef, "buildrun ref must not be nil") {
					return
				}
				oneShotRef := *maybeOneShotRef

				spec := optest.MustGetSpec[buildrun.AlphaSpec1](plaid, oneShotRef)
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
				creationStatus.Wait(t, ctx)
				projectStatus := optest.MustGetStatus[Alpha1Status](plaid, projectRef)
				maybeOneShotRef := projectStatus.OneShots[0].Ref
				if !assert.NotNil(t, maybeOneShotRef, "buildrun ref must not be nil") {
					return
				}
				oneShotRef := *maybeOneShotRef

				oneShotFinished := projectWatcher.Status.Fork()
				MustFinishBuildRunRun(t, ctx, plaid.Legacy.Store, oneShotRef, 0)

				t.Run("Then the exit is noted in the project status", func(t *testing.T) {
					oneShotFinished.WaitFor(t, ctx, func(ctx context.Context) bool {
						status := optest.MustGetStatus[Alpha1Status](plaid, projectRef)
						return status.Ready
					})
					status := optest.MustGetStatus[Alpha1Status](plaid, projectRef)
					if assert.Len(t, status.OneShots, 1) {
						assert.True(t, status.OneShots[0].Done, "must have completed")
					}

					assert.True(t, status.Done, "project is completed.")
				})
			})
		})
	})
}
