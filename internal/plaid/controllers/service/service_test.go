package service

import (
	"context"
	"github.com/meschbach/plaid/internal/junk"
	"github.com/meschbach/plaid/internal/plaid/controllers/dependencies"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestService(t *testing.T) {
	ctx, done := context.WithCancel(context.Background())
	t.Cleanup(done)

	core := resources.WithTestSubsystem(t, ctx)
	core.AttachController("controller.service", NewSystem(core.Controller))
	store := core.Store
	watcher, err := store.Watcher(ctx)
	core.AttachController("test.watcher", watcher)
	require.NoError(t, err)

	t.Run("Given an new Service with no build or dependencies", func(t *testing.T) {
		serviceRef := resources.FakeMetaOf(Alpha1)
		subject, err := Observe(ctx, watcher, serviceRef)
		require.NoError(t, err)
		serviceStatusUpdate := subject.Status

		spec := Alpha1Spec{
			Run: exec.TemplateAlpha1Spec{
				Command:    "test-command",
				WorkingDir: "/knight",
			},
		}

		createStatusChange := serviceStatusUpdate.Fork()
		require.NoError(t, store.Create(ctx, serviceRef, spec))

		t.Run("Then a new command should be created", func(t *testing.T) {
			createStatusChange.Wait()
			var status Alpha1Status
			exists, err := store.GetStatus(ctx, serviceRef, &status)
			require.NoError(t, err)
			assert.True(t, exists, "must exist")
			assert.False(t, status.Ready, "must not be ready %#v", status)
			assert.Equal(t, Running, status.Run.State, "must be in a running state")
			assert.NotNil(t, status.Run.Ref, "must reference a run")

			t.Run("When the process has started", func(t *testing.T) {
				runRef := *status.Run.Ref
				procRun := serviceStatusUpdate.Fork()
				now := time.Now()
				exists, err := store.UpdateStatus(ctx, runRef, exec.InvocationAlphaV1Status{
					Started: &now,
					Healthy: true,
				})
				require.NoError(t, err)
				require.True(t, exists, "invocation must exit")
				procRun.Wait()

				var status Alpha1Status
				exists, err = store.GetStatus(ctx, serviceRef, &status)
				require.NoError(t, err)
				assert.True(t, exists, "must exist")
				assert.True(t, status.Ready, "should be ready now %#v", status)
				assert.Equal(t, Running, status.Run.State, "must be in a running state")
				assert.NotNil(t, status.Run.Ref, "must reference a run")
			})
		})
	})

	t.Run("Given a new service with a dependency", func(t *testing.T) {
		serviceRef := resources.FakeMetaOf(Alpha1)
		subject, err := Observe(ctx, watcher, serviceRef)
		require.NoError(t, err)
		serviceStatusUpdate := subject.Status

		spec := Alpha1Spec{
			Dependencies: []resources.Meta{
				resources.FakeMeta(),
			},
			Run: exec.TemplateAlpha1Spec{
				Command:    "test-command",
				WorkingDir: "/knight",
			},
		}

		createStatusChange := serviceStatusUpdate.Fork()
		require.NoError(t, store.Create(ctx, serviceRef, spec))

		t.Run("When the dependency does not exist", func(t *testing.T) {
			createStatusChange.Wait()
			var status Alpha1Status
			exists, err := store.GetStatus(ctx, serviceRef, &status)
			require.NoError(t, err)
			assert.True(t, exists, "must exist")
			assert.False(t, status.Ready, "must not be ready %#v", status)
			if assert.Len(t, status.Dependencies, 1) {
				assert.False(t, status.Dependencies[0].Ready)
				assert.Equal(t, spec.Dependencies[0], status.Dependencies[0].Dependency)
			}
			assert.Equal(t, "not-ready", status.Run.State, "then must be not ready")
			assert.Nil(t, status.Run.Ref, "no run reference created")

			t.Run("And the dependency is created", func(t *testing.T) {
				require.NoError(t, store.Create(ctx, spec.Dependencies[0], dependencies.ReadyAlpha1Status{Ready: true}))

				var status Alpha1Status
				exists, err := store.GetStatus(ctx, serviceRef, &status)
				require.NoError(t, err)
				assert.True(t, exists, "service must still exist")
				assert.Nil(t, status.Run.Ref, "run should not be created")
			})

			t.Run("And the dependency becomes ready", func(t *testing.T) {
				dependencyReady := subject.Status.Fork()
				exists, err := store.UpdateStatus(ctx, spec.Dependencies[0], dependencies.ReadyAlpha1Status{Ready: true})
				require.NoError(t, err)
				require.True(t, exists)

				dependencyReady.Wait()
				var status Alpha1Status
				exists, err = store.GetStatus(ctx, serviceRef, &status)
				require.NoError(t, err)
				assert.True(t, exists, "service must still exist")
				if assert.Len(t, status.Dependencies, 1) {
					assert.True(t, status.Dependencies[0].Ready, "dep must be ready")
				}
				assert.Equal(t, Running, status.Run.State, "must have started run")
				assert.NotNil(t, status.Run.Ref, "run reference must be provided")
			})
		})
	})
}

type WatchedResource struct {
	Changes *junk.ChangeTracker
	Status  *junk.ChangeTracker
}

func Observe(ctx context.Context, watcher *resources.ClientWatcher, ref resources.Meta) (*WatchedResource, error) {
	w := &WatchedResource{
		Changes: junk.NewChangeTracker(),
		Status:  junk.NewChangeTracker(),
	}
	_, err := watcher.OnResource(ctx, ref, func(ctx context.Context, changed resources.ResourceChanged) error {
		w.Changes.Update()
		switch changed.Operation {
		case resources.StatusUpdated:
			w.Status.Update()
		default:
		}
		return nil
	})
	return w, err
}
