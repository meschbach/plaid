package alpha2

import (
	"context"
	"github.com/go-faker/faker/v4"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/internal/plaid/controllers/probes"
	"github.com/meschbach/plaid/internal/plaid/httpProbe"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/optest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestProbeReadiness(t *testing.T) {
	_, plaid := optest.New(t)
	//todo: System should become a top level element
	serviceController := NewController(plaid.Legacy.System)
	//todo: revisit how the supervision tree is accessed
	plaid.Legacy.AttachController("service/alpha2", serviceController)

	plaid.Run("Given a running service", func(t *testing.T, plaid *optest.System, ctx context.Context) {
		ref := resources.FakeMetaOf(Type)
		refWatch := plaid.Observe(ctx, ref)
		initToken := faker.Word()
		simpleSpec := Spec{
			Run: exec.TemplateAlpha1Spec{
				Command:    "echo test",
				WorkingDir: "/tmp/systest",
			},
			Readiness: &probes.TemplateAlpha1Spec{
				Http: &httpProbe.TemplateAlpha1{
					Port: 9999,
					Path: "/",
				},
			},
			RestartToken: initToken,
		}

		createStatusGate := refWatch.Status.Fork()
		plaid.MustCreate(ctx, ref, simpleSpec)
		createStatusGate.Wait(t, ctx, "failed to update status on create")

		status := optest.MustGetStatus[Status](plaid, ref)
		require.Nil(t, status.Stable, "build must not be promoted")
		require.NotNil(t, status.Next, "must be working on build")
		require.NotNil(t, status.Next.Service, "run proc must not be nil")
		assert.Nil(t, status.Next.Probe.Ref, "probe should not have been created yet")

		startedGate := refWatch.Status.Fork()
		now := time.Now()
		plaid.MustUpdateStatus(ctx, *status.Next.Service, exec.InvocationAlphaV1Status{
			Started: &now,
			Healthy: true,
		})
		startedGate.Wait(t, ctx)

		status = optest.MustGetStatus[Status](plaid, ref)
		assert.False(t, status.Ready, "probe gate not ready, so service should not be ready")
		assert.Nil(t, status.Stable, "no stable service exists")
		if assert.NotNil(t, status.Next, "still working on service") {
			if assert.NotNil(t, status.Next.Probe.Ref, "probe ref should not be nil") {
				probeRef := *status.Next.Probe.Ref

				plaid.Run("When the probe shows ready", func(t *testing.T, plaid *optest.System, ctx context.Context) {
					probeReadyGate := refWatch.Status.Fork()
					plaid.MustUpdateStatus(ctx, probeRef, httpProbe.AlphaV1Status{Ready: true})
					probeReadyGate.Wait(t, ctx, "updated probe should result in status change")

					afterProbeReady := optest.MustGetStatus[Status](plaid, ref)
					assert.Nil(t, afterProbeReady.Next, "promoted build")
					if assert.NotNil(t, afterProbeReady.Stable, "promoted build") {
					}
					assert.True(t, afterProbeReady.Ready, "service should be ready")
				})
			}
		}

		plaid.Run("When deleting the service", func(t *testing.T, plaid *optest.System, ctx context.Context) {
			beforeDelete := optest.MustGetStatus[Status](plaid, ref)

			serviceDeleted := refWatch.Delete.Fork()
			if assert.NotNil(t, beforeDelete.Stable, "stable must exist to delete") {
				stableDelete := plaid.Observe(ctx, *beforeDelete.Stable.Service).Delete.Fork()
				plaid.MustDelete(ctx, ref)
				serviceDeleted.Wait(t, ctx, "service must be deleted")
				stableDelete.Wait(t, ctx, "stable should be deleted")
			}
		})
	})
}
