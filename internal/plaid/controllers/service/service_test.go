package service

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/controllers/dependencies"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/optest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestServiceWithNoDependencies(t *testing.T) {
	ctx, plaid := optest.New(t)
	core := plaid.Legacy
	core.AttachController("controller.service", NewSystem(core.Controller))
	store := core.Store

	t.Run("Given an new Service with no build or dependencies", func(t *testing.T) {
		serviceRef := resources.FakeMetaOf(Alpha1)
		subject := plaid.Observe(ctx, serviceRef)
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
			createStatusChange.Wait(ctx)
			status := optest.MustGetStatus[Alpha1Status](plaid, serviceRef)
			assert.False(t, status.Ready, "must not be ready %#v", status)
			assert.Equal(t, Running, status.Run.State, "must be in a running state")
			assert.NotNil(t, status.Run.Ref, "must reference a run")

			t.Run("When the process has started", func(t *testing.T) {
				runRef := *status.Run.Ref
				now := time.Now()
				invocationFinished := serviceStatusUpdate.Fork()
				optest.MustUpdateStatusAndWait(plaid, serviceStatusUpdate, runRef, exec.InvocationAlphaV1Status{
					Started: &now,
					Healthy: true,
				})

				invocationFinished.WaitFor(ctx, func(ctx context.Context) bool {
					status := optest.MustGetStatus[Alpha1Status](plaid, serviceRef)
					return status.Ready
				})
				status := optest.MustGetStatus[Alpha1Status](plaid, serviceRef)
				assert.True(t, status.Ready, "should be ready now %#v", status)
				assert.Equal(t, Running, status.Run.State, "must be in a running state")
				assert.NotNil(t, status.Run.Ref, "must reference a run")
			})
		})
	})
}

func TestServiceWithDependencies(t *testing.T) {
	ctx, plaid := optest.New(t)
	core := plaid.Legacy
	core.AttachController("controller.service", NewSystem(core.Controller))
	store := core.Store
	t.Run("Given a new service with a dependency", func(t *testing.T) {
		serviceRef := resources.FakeMetaOf(Alpha1)
		subject := plaid.Observe(ctx, serviceRef)
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
			createStatusChange.Wait(ctx)
			status := optest.MustGetStatus[Alpha1Status](plaid, serviceRef)
			assert.False(t, status.Ready, "must not be ready %#v", status)
			if assert.Len(t, status.Dependencies, 1) {
				assert.False(t, status.Dependencies[0].Ready)
				assert.Equal(t, spec.Dependencies[0], status.Dependencies[0].Dependency)
			}
			assert.Equal(t, "not-ready", status.Run.State, "then must be not ready")
			assert.Nil(t, status.Run.Ref, "no run reference created")

			t.Run("And the dependency is created", func(t *testing.T) {
				require.NoError(t, store.Create(ctx, spec.Dependencies[0], dependencies.ReadyAlpha1Status{Ready: false}))

				status := optest.MustGetStatus[Alpha1Status](plaid, serviceRef)
				assert.Nil(t, status.Run.Ref, "run should not be created")
			})

			t.Run("And the dependency becomes ready", func(t *testing.T) {
				optest.MustUpdateStatusAndWait(plaid, subject.Status, spec.Dependencies[0], dependencies.ReadyAlpha1Status{Ready: true})

				status := optest.MustGetStatus[Alpha1Status](plaid, serviceRef)
				if assert.Len(t, status.Dependencies, 1) {
					assert.True(t, status.Dependencies[0].Ready, "dep must be ready")
				}
				assert.Equal(t, Running, status.Run.State, "must have started run")
				assert.NotNil(t, status.Run.Ref, "run reference must be provided")
			})
		})
	})
}
