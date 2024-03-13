package alpha2

import (
	"context"
	"github.com/go-faker/faker/v4"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/optest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSimpleLifecycle(t *testing.T) {
	_, plaid := optest.New(t)
	//todo: System should become a top level element
	serviceController := NewController(plaid.Legacy.System)
	//todo: revisit how the supervision tree is accessed
	plaid.Legacy.AttachController("service/alpha2", serviceController)

	plaid.Run("When creating a new service as just a running process", func(t *testing.T, plaid *optest.System, ctx context.Context) {
		start := time.Now()
		ref := resources.FakeMetaOf(Type)
		refWatch := plaid.Observe(ctx, ref)
		initToken := ""
		simpleSpec := Spec{
			Run: exec.TemplateAlpha1Spec{
				Command:    "echo test",
				WorkingDir: "/tmp/systest",
			},
			RestartToken: initToken,
		}

		createStatusGate := refWatch.Status.Fork()
		plaid.MustCreate(ctx, ref, simpleSpec)
		createStatusGate.Wait(t, ctx, "failed to update status on create")

		status := optest.MustGetStatus[Status](plaid, ref)
		if assert.NotNil(t, status.Next, "attempting to build") {
			assert.Equal(t, simpleSpec.RestartToken, status.Next.Token, "build token should be the same")
			assert.Less(t, start, status.Next.Last, "last modified should be less than start time")
			assert.Equal(t, initToken, status.Next.Token, "token is correctly marked")
			assert.NotNil(t, status.Next.Service, "a service should be created")
		}

		assert.Nil(t, status.Stable, "no stable build exists")

		if status.Next.Service == nil {
			return
		}

		plaid.Run("When the invocation starts starts", func(t *testing.T, plaid *optest.System, ctx context.Context) {
			statusChange := refWatch.Status.Fork()
			now := time.Now()
			plaid.MustUpdateStatus(ctx, *status.Next.Service, exec.InvocationAlphaV1Status{
				Started: &now,
				Healthy: true,
			})
			statusChange.Wait(t, ctx, "service should update")

			afterStartStatus := optest.MustGetStatus[Status](plaid, ref)
			assert.Nil(t, afterStartStatus.Next)
			if assert.NotNil(t, afterStartStatus.Stable, "next should be promoted to stable: %#v", afterStartStatus) {
				assert.Less(t, now, afterStartStatus.Stable.Last, "update time is after the service started")
				assert.Equal(t, initToken, afterStartStatus.Stable.Token, "token is correct")
				assert.Equal(t, *status.Next.Service, *afterStartStatus.Stable.Service, "stable and old next should reference the same process")
			}
		})

		plaid.Run("When the restart token is changed", func(t *testing.T, plaid *optest.System, ctx context.Context) {
			nextTokenStart := time.Now()
			nextToken := faker.Word()
			simpleSpec.RestartToken = nextToken

			plaid.MustUpdateAndWait(refWatch.Status, ref, simpleSpec)
			statusOnRestart := optest.MustGetStatus[Status](plaid, ref)

			if assert.NotNil(t, statusOnRestart.Stable, "stable remains unchanged") {
				assert.Equal(t, initToken, statusOnRestart.Stable.Token)
			}

			if assert.NotNil(t, statusOnRestart.Next, "a new build should be started") {
				assert.Equal(t, nextToken, statusOnRestart.Next.Token)
				assert.Less(t, nextTokenStart, statusOnRestart.Next.Last)
				assert.Empty(t, statusOnRestart.Old, "no old build yet")

				plaid.Run("And the build interrupted", func(t *testing.T, plaid *optest.System, ctx context.Context) {
					interruptTime := time.Now()
					interruptingToken := faker.Word()

					simpleSpec.RestartToken = interruptingToken
					plaid.MustUpdateAndWait(refWatch.Status, ref, simpleSpec)
					interruptedStatus := optest.MustGetStatus[Status](plaid, ref)

					if assert.NotNil(t, interruptedStatus.Stable, "stable remains unchanged") {
						assert.Equal(t, initToken, statusOnRestart.Stable.Token)
					}

					if assert.NotNil(t, interruptedStatus.Next) {
						assert.Equal(t, interruptingToken, interruptedStatus.Next.Token, "next token should be the interrupted value")
						assert.Less(t, interruptTime, interruptedStatus.Next.Last, "last change time should be updated")
					}

					optest.MustBeMissingSpec[exec.InvocationAlphaV1Spec](plaid, *statusOnRestart.Next.Service, "old build is missing")
					if assert.Len(t, interruptedStatus.Old, 1, "interrupted build prompted to 'old'") {
						assert.Equal(t, *statusOnRestart.Next.Service, *interruptedStatus.Old[0].Service)
					}

					plaid.Run("And the build completed", func(t *testing.T, plaid *optest.System, ctx context.Context) {
						buildCompletion := refWatch.Status.Fork()
						now := time.Now()
						plaid.MustUpdateStatus(ctx, *interruptedStatus.Next.Service, exec.InvocationAlphaV1Status{
							Started: &now,
							Healthy: true,
						})

						buildCompletion.Wait(t, ctx, "new build should be promoted")
						afterBuild := optest.MustGetStatus[Status](plaid, ref)

						assert.Nil(t, afterBuild.Next, "build has completed")
						assert.Len(t, afterBuild.Old, 2, "interrupted build and old service are retired")
						assert.Equal(t, interruptingToken, afterBuild.Stable.Token, "interrupted build was promoted")

						for _, old := range afterBuild.Old {
							var oldSpec exec.InvocationAlphaV1Spec
							exists, err := plaid.Legacy.Store.Get(ctx, *old.Service, &oldSpec)
							require.NoError(t, err)
							assert.False(t, exists, "old builds should no longer exist, %s did though\n\tservice: %s\n\tinterrupted: %s", *old.Service, *statusOnRestart.Stable.Service, *statusOnRestart.Next.Service)
						}
					})
				})
			}
		})
	})
}
