package project

import (
	"context"
	"github.com/go-faker/faker/v4"
	"github.com/meschbach/plaid/controllers/service/alpha2"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/internal/plaid/controllers/service"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/optest"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestProjectAlpha1(t *testing.T) {
	t.Run("Given a Plaid instance with the configured controller", func(t *testing.T) {
		_, plaid := optest.New(t)
		plaid.Legacy.AttachController("plaid.controllers.project", NewProjectSystem(plaid.Legacy.Controller))

		plaid.Run("When a new project is Created with a daemon service", func(t *testing.T, s *optest.System, ctx context.Context) {
			tmpDir := os.TempDir()
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
			plaid.MustCreate(ctx, projectRef, projectSpec)

			plaid.Run("Then the initial status is setup", func(t *testing.T, plaid *optest.System, ctx context.Context) {
				createStatusChange.Wait(t, ctx)
				status := optest.MustGetStatus[Alpha1Status](plaid, projectRef)
				assert.False(t, status.Ready, "initial status must not be ready")
			})

			plaid.Run("Then a service is Created", func(t *testing.T, s *optest.System, ctx context.Context) {
				createStatusChange.Wait(t, ctx)

				status := optest.MustGetStatus[Alpha1Status](plaid, projectRef)
				maybeServiceRef := status.Daemons[0].Current
				if !assert.NotNil(t, maybeServiceRef, "daemon current reference") {
					return
				}
				serviceRef := *maybeServiceRef
				spec := optest.MustGetSpec[service.Alpha1Spec](plaid, serviceRef)

				plaid.Run("with the correct build configuration", func(t *testing.T, s *optest.System, ctx context.Context) {
					assert.Equal(t, tmpDir, spec.Build.WorkingDir, "working directory set correctly")
					assert.Equal(t, builderCommand, spec.Build.Command, "builder command set correctly")
				})
				plaid.Run("with the correct run configuration", func(t *testing.T, s *optest.System, ctx context.Context) {
					assert.Equal(t, tmpDir, spec.Run.WorkingDir, "working directory set correctly")
					assert.Equal(t, runCommand, spec.Run.Command, "run command set correctly")
				})

				plaid.Run("Then the service is not ready", func(t *testing.T, s *optest.System, ctx context.Context) {
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

				plaid.Run("When the service is ready", func(t *testing.T, s *optest.System, ctx context.Context) {
					serviceUpdated := projectWatch.Status.Fork()
					optest.MustUpdateStatusAndWait(plaid, projectWatch.Status, serviceRef, alpha2.Status{
						LatestToken: "",
						Stable: &alpha2.TokenStatus{
							Token: "",
						},
						//todo: these should probably be in TokenStatus
						//Dependencies: nil,
						//Build: service.Alpha1BuildStatus{
						//	State: Alpha1StateSuccess,
						//},
						//Ready: true,
					})

					serviceUpdated.WaitFor(t, ctx, func(ctx context.Context) bool {
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
