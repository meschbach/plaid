package service

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/optest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestServiceWithBuilderStep(t *testing.T) {
	ctx, done := context.WithCancel(context.Background())
	t.Cleanup(done)

	ctx, plaid := optest.New(t)
	core := plaid.Legacy
	core.AttachController("controller.service", NewSystem(core.Controller))
	store := core.Store

	t.Run("Given a new service with a build step and no dependencies", func(t *testing.T) {
		serviceRef := resources.FakeMetaOf(Alpha1)
		subject := plaid.Observe(ctx, serviceRef)

		buildSpec := exec.TemplateAlpha1Spec{
			Command:    "build-step",
			WorkingDir: "/on/this/hill",
		}
		spec := Alpha1Spec{
			Build: &buildSpec,
			Run:   exec.TemplateAlpha1Spec{},
		}

		statusCreateChange := subject.Status.Fork()
		require.NoError(t, store.Create(ctx, serviceRef, spec))

		t.Run("When the service is created created", func(t *testing.T) {
			statusCreateChange.Wait(ctx)
			status := optest.MustGetStatus[Alpha1Status](plaid, serviceRef)
			assert.Equal(t, "starting", status.Build.State, "then the builder must be created")
			require.NotNil(t, status.Build.Ref, "then the build must select the reference")

			assertRunNotReady(t, status)

			builderRef := *status.Build.Ref
			startTime := time.Now()
			t.Run("And the build is started", func(t *testing.T) {
				started := subject.Status.Fork()
				optest.MustUpdateStatusAndWait(plaid, subject.Status, builderRef, exec.InvocationAlphaV1Status{
					Started: &startTime,
				})

				started.WaitFor(ctx, func(ctx context.Context) bool {
					status := optest.MustGetStatus[Alpha1Status](plaid, serviceRef)
					return status.Build.State == "running"
				})
				status := optest.MustGetStatus[Alpha1Status](plaid, serviceRef)
				assert.Equal(t, builderRef, *status.Build.Ref, "ref must be maintained")
				assert.Equal(t, "running", status.Build.State, "then the state must be running")

				assertRunNotReady(t, status)
			})

			finishedTime := time.Now()
			t.Run("And the build completes successfully", func(t *testing.T) {
				finished := subject.Status.Fork()
				exitStatus := 0
				optest.MustUpdateStatusAndWait(plaid, subject.Status, builderRef, exec.InvocationAlphaV1Status{
					Started:    &startTime,
					Finished:   &finishedTime,
					ExitStatus: &exitStatus,
					Healthy:    true,
				})

				finished.WaitFor(ctx, func(ctx context.Context) bool {
					status := optest.MustGetStatus[Alpha1Status](plaid, serviceRef)
					return status.Build.State == "finished"
				})
				status := optest.MustGetStatus[Alpha1Status](plaid, serviceRef)
				assert.Equal(t, builderRef, *status.Build.Ref, "ref must be maintained")
				assert.Equal(t, "finished", status.Build.State, "then the state must be finished")

				t.Run("Then the service should be started", func(t *testing.T) {
					assert.NotEqual(t, StateNotReady, status.Run.State, "then the service should be creating")
					assert.NotNil(t, status.Run.Ref, "then the service run ref should be created")
				})
			})
		})
	})
}

func assertRunNotReady(t *testing.T, status Alpha1Status) {
	assert.Nil(t, status.Run.Ref, "then the run should not be created")
	assert.Equal(t, StateNotReady, status.Run.State, "then the service run state should not be ready")
}
