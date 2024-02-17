package project

import (
	"context"
	"github.com/go-faker/faker/v4"
	"github.com/meschbach/plaid/internal/junk"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/internal/plaid/controllers/service"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/optest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestProjectAlpha1(t *testing.T) {
	t.Run("Given a Plaid instance with the configured controller", func(t *testing.T) {
		ctx, plaid := optest.New(t)
		plaid.Legacy.AttachController("plaid.controllers.project", NewProjectSystem(plaid.Legacy.Controller))

		t.Run("When a new project is Created with a daemon service", func(t *testing.T) {
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
			projectWatch := plaid.Observe(ctx, projectRef)

			createStatusChange := projectWatch.Status.Fork()
			require.NoError(t, plaid.Legacy.Store.Create(ctx, projectRef, projectSpec))

			t.Run("Then the initial status is setup", func(t *testing.T) {
				createStatusChange.Wait(ctx)
				status := optest.MustGetStatus[Alpha1Status](plaid, projectRef)
				assert.False(t, status.Ready, "initial status must not be ready")
			})

			t.Run("Then a service is Created", func(t *testing.T) {
				createStatusChange.Wait(ctx)
				status := optest.MustGetStatus[Alpha1Status](plaid, projectRef)
				maybeServiceRef := status.Daemons[0].Current
				if !assert.NotNil(t, maybeServiceRef, "daemon current reference") {
					return
				}
				serviceRef := *maybeServiceRef
				spec := optest.MustGetSpec[service.Alpha1Spec](plaid, serviceRef)

				t.Run("with the correct build configuration", func(t *testing.T) {
					assert.Equal(t, tmpDir, spec.Build.WorkingDir, "working directory set correctly")
					assert.Equal(t, builderCommand, spec.Build.Command, "builder command set correctly")
				})
				t.Run("with the correct run configuration", func(t *testing.T) {
					assert.Equal(t, tmpDir, spec.Run.WorkingDir, "working directory set correctly")
					assert.Equal(t, runCommand, spec.Run.Command, "run command set correctly")
				})

				t.Run("Then the service is not ready", func(t *testing.T) {
					optest.MustUpdateStatus(plaid, serviceRef, service.Alpha1Status{
						Dependencies: nil,
						Build: service.Alpha1BuildStatus{
							State: Alpha1StateSuccess,
						},
						Ready: false,
					})

					projectStatus := optest.MustGetStatus[Alpha1Status](plaid, projectRef)
					assert.False(t, projectStatus.Ready, "service is not ready")
				})

				t.Run("When the service is ready", func(t *testing.T) {
					serviceUpdated := projectWatch.Status.Fork()
					optest.MustUpdateStatusAndWait(plaid, projectWatch.Status, serviceRef, service.Alpha1Status{
						Dependencies: nil,
						Build: service.Alpha1BuildStatus{
							State: Alpha1StateSuccess,
						},
						Ready: true,
					})

					serviceUpdated.WaitFor(ctx, func(ctx context.Context) bool {
						projectStatus := optest.MustGetStatus[Alpha1Status](plaid, projectRef)
						return projectStatus.Ready
					})
					projectStatus := optest.MustGetStatus[Alpha1Status](plaid, projectRef)
					assert.True(t, projectStatus.Ready, "service should be ready, got %#v\n", projectStatus)
				})
			})
		})
	})
}
