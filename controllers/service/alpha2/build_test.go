package alpha2

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/optest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuild(t *testing.T) {
	_, plaid := optest.New(t)
	//todo: System should become a top level element
	serviceController := NewController(plaid.Legacy.System)
	//todo: revisit how the supervision tree is accessed
	plaid.Legacy.AttachController("service/alpha2", serviceController)

	plaid.Run("Given a service with a build", func(t *testing.T, plaid *optest.System, ctx context.Context) {
		ref := resources.FakeMetaOf(Type)
		refWatcher := plaid.Observe(ctx, ref)

		initSpec := Spec{
			Build: &exec.TemplateAlpha1Spec{
				Command:    "echo test",
				WorkingDir: "/some/location",
			},
			Run: exec.TemplateAlpha1Spec{
				Command:    "echo kite",
				WorkingDir: "/other/location",
			},
		}

		createChange := refWatcher.Status.Fork()
		plaid.MustCreate(ctx, ref, initSpec)
		createChange.Wait(t, ctx, "pickup created spec")

		status := optest.MustGetStatus[Status](plaid, ref)
		if assert.NotNil(t, status.Next.Build, "build is created") {

		}
	})
}
