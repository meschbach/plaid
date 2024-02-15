package service

import (
	"context"
	"github.com/meschbach/plaid/internal/junk"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestServiceWithBuilderStep(t *testing.T) {
	ctx, done := context.WithCancel(context.Background())
	t.Cleanup(done)

	core := resources.WithTestSubsystem(t, ctx)
	core.AttachController("controller.service", NewSystem(core.Controller))
	store := core.Store
	watcher, err := store.Watcher(ctx)
	core.AttachController("test.watcher", watcher)
	require.NoError(t, err)

	t.Run("Given a new service with a build step and no dependencies", func(t *testing.T) {
		serviceRef := resources.FakeMetaOf(Alpha1)
		subject, err := Observe(ctx, watcher, serviceRef)
		require.NoError(t, err)

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
			statusCreateChange.Wait()
			status := MustGetStatus[Alpha1Status](t, ctx, store, serviceRef)
			assert.Equal(t, "create", status.Build.State, "then the builder must be created")
			require.NotNil(t, status.Build.Ref, "then the build must select the reference")

			assertRunNotReady(t, status)

			builderRef := *status.Build.Ref
			startTime := time.Now()
			t.Run("And the build is started", func(t *testing.T) {
				MustUpdateStatusAndWait(t, ctx, store, subject.Status, builderRef, exec.InvocationAlphaV1Status{
					Started: &startTime,
				})

				status := MustGetStatus[Alpha1Status](t, ctx, store, serviceRef)
				assert.Equal(t, builderRef, *status.Build.Ref, "ref must be maintained")
				assert.Equal(t, "running", status.Build.State, "then the state must be starting")

				assertRunNotReady(t, status)
			})

			finishedTime := time.Now()
			t.Run("And the build completes successfully", func(t *testing.T) {
				exitStatus := 0
				MustUpdateStatusAndWait(t, ctx, store, subject.Status, builderRef, exec.InvocationAlphaV1Status{
					Started:    &startTime,
					Finished:   &finishedTime,
					ExitStatus: &exitStatus,
					Healthy:    true,
				})

				status := MustGetStatus[Alpha1Status](t, ctx, store, serviceRef)
				assert.Equal(t, builderRef, *status.Build.Ref, "ref must be maintained")
				assert.Equal(t, "finished", status.Build.State, "then the state must be starting")

				t.Run("Then the service should be started", func(t *testing.T) {
					assert.NotEqual(t, StateNotReady, status.Run.State, "then the service should be creating")
					assert.NotNil(t, status.Run.Ref, "then the service run ref should be created")
				})
			})
		})
	})
}

func MustUpdateStatusAndWait(t *testing.T, ctx context.Context, store *resources.Client, gate *junk.ChangeTracker, ref resources.Meta, status any) {
	postUpdate := gate.Fork()
	exists, err := store.UpdateStatus(ctx, ref, status)
	require.NoError(t, err, "update must not error")
	require.True(t, exists, "build must exist")

	postUpdate.Wait()
}

func MustGetStatus[Status any](t *testing.T, ctx context.Context, store *resources.Client, ref resources.Meta) Status {
	var out Status
	exists, err := store.GetStatus(ctx, ref, &out)
	require.NoError(t, err, "failed to retrieve status of %s", ref)
	require.True(t, exists, "resource %s was expected to exist and status retrieval but was not", ref)
	return out
}

func assertRunNotReady(t *testing.T, status Alpha1Status) {
	assert.Nil(t, status.Run.Ref, "then the run should not be created")
	assert.Equal(t, StateNotReady, status.Run.State, "then the service run state should not be ready")
}
