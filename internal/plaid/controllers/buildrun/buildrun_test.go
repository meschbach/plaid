package buildrun

import (
	"context"
	"git.meschbach.com/mee/platform.git/plaid/internal/junk"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/controllers/exec"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestBuildRunControllerAlpha1(t *testing.T) {
	onCleanupTracing := junk.SetupTestTracing(t)
	t.Cleanup(func() {
		ctx, done := context.WithTimeout(context.Background(), 1*time.Second)
		defer done()

		onCleanupTracing(ctx)
	})
	baseContext, done := context.WithCancel(context.Background())
	t.Cleanup(done)

	t.Run("Given a new Alpha1 Resource", func(t *testing.T) {
		ctx, _ := junk.TraceSubtest(t, baseContext, tracer)
		core := resources.WithTestSubsystem(t, ctx)

		core.AttachController("controller.buildrun", &Controller{storage: core.Controller})

		exampleRestartToken := "copper wire"
		ref := resources.FakeMetaOf(Alpha1)
		spec := AlphaSpec1{
			RestartToken: exampleRestartToken,
			Build: exec.TemplateAlpha1Spec{
				Command: "echo test",
			},
			Run: exec.TemplateAlpha1Spec{
				Command: "fly away",
			},
		}
		dataPlane := core.Store
		require.NoError(t, dataPlane.Create(ctx, ref, spec))
		var buildRunState AlphaStatus1
		success, err := resources.WaitOn(ctx, resources.ForStatusState(core.Store, ref, &buildRunState, func(status *AlphaStatus1) (bool, error) {
			return buildRunState.Build.Ref != nil, nil
		}))
		require.NoError(t, err)
		require.True(t, success)

		t.Run("When the initial build has not completed", func(t *testing.T) {
			ctx, _ = junk.TraceSubtest(t, ctx, tracer)

			t.Run("Then the build token is the one specified", func(t *testing.T) {
				ctx, _ = junk.TraceSubtest(t, ctx, tracer)
				var status AlphaStatus1
				completed, err := resources.WaitOn(ctx, resources.ForStatusState[AlphaStatus1](dataPlane, ref, &status, func(status *AlphaStatus1) (bool, error) {
					return status.Build.Token == exampleRestartToken, nil
				}))
				require.NoError(t, err, "Status: %#v\n", status)
				assert.True(t, completed, "eventually converges on the restart token.")
			})
		})

		t.Run("When the initial build completed", func(t *testing.T) {
			ctx, _ = junk.TraceSubtest(t, ctx, tracer)
			now := time.Now()
			exitCode := 0
			exists, err := core.Store.UpdateStatus(ctx, *buildRunState.Build.Ref, exec.InvocationAlphaV1Status{
				Started:    &now,
				Finished:   &now,
				ExitStatus: &exitCode,
				Healthy:    true,
			})
			require.NoError(t, err)
			require.True(t, exists, "build must exist")

			t.Run("Then the run token is updated", func(t *testing.T) {
				ctx, _ = junk.TraceSubtest(t, ctx, tracer)
				originalToken := buildRunState.Run.Token
				success, err := resources.WaitOn(ctx, resources.ForStatusState(core.Store, ref, &buildRunState, func(status *AlphaStatus1) (bool, error) {
					return status.Run.Token != originalToken, nil
				}))
				require.NoError(t, err)
				require.True(t, success)
			})
		})
	})
}
