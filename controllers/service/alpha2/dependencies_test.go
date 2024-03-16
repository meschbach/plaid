package alpha2

import (
	"context"
	"github.com/go-faker/faker/v4"
	"github.com/meschbach/plaid/internal/plaid/controllers/dependencies"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/optest"
	"github.com/stretchr/testify/assert"
	"testing"
)

type ready struct {
	Ready bool `json:"ready"`
}

func TestDependencies(t *testing.T) {
	_, plaid := optest.New(t)
	//todo: System should become a top level element
	serviceController := NewController(plaid.Legacy.System)
	//todo: revisit how the supervision tree is accessed
	plaid.Legacy.AttachController("service/alpha2", serviceController)

	plaid.Run("Given a service with a dependency which does not exist", func(t *testing.T, plaid *optest.System, ctx context.Context) {
		depTarget := resources.FakeMeta()
		depName := faker.Word()
		initSpec := Spec{
			Dependencies: dependencies.Alpha1Spec{
				dependencies.NamedDependencyAlpha1{
					Name: depName,
					Ref:  depTarget,
				},
			},
			Run: exec.TemplateAlpha1Spec{
				Command:    "echo test",
				WorkingDir: "/example",
			},
			RestartToken: faker.Word(),
		}

		ref := resources.FakeMetaOf(Type)
		refWatcher := plaid.Observe(ctx, ref)

		createStatusGate := refWatcher.Status.Fork()
		plaid.MustCreate(ctx, ref, initSpec)
		createStatusGate.Wait(t, ctx)

		initStatus := optest.MustGetStatus[Status](plaid, ref)
		assert.False(t, initStatus.Ready, "service is not ready")
		if assert.NotNil(t, initStatus.Next, "next build has began") {
			assert.Nil(t, initStatus.Next.Service, "service should not have been created")
			assert.Equal(t, TokenStageDependencyWait, initStatus.Next.Stage, "then the build is waiting on dependencies to be ready")
			if assert.Len(t, initStatus.Next.Deps, 1, "dependency status is listed") {
				assert.False(t, initStatus.Next.Deps[depName].Ready)
			}
		}

		plaid.Run("When the dependency gets created with a status", func(t *testing.T, plaid *optest.System, ctx context.Context) {
			statusChangeGate := refWatcher.Status.Fork()
			plaid.MustCreate(ctx, depTarget, ready{Ready: true})
			plaid.MustUpdateStatus(ctx, depTarget, ready{Ready: true})
			statusChangeGate.Wait(t, ctx)

			depReady := optest.MustGetStatus[Status](plaid, ref)
			if assert.NotNil(t, depReady.Next, "service transitioned to stable") {
				assert.True(t, depReady.Next.DepsFuse, "all dependencies were ready at once")
				assert.True(t, depReady.Next.Deps[depName].Ready, "dependency is ready")
			}
		})
	})
}
