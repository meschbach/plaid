package local

import (
	"context"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/controllers/exec"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/controllers/logdrain"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources/operator"
	"github.com/thejerf/suture/v4"
	"go.opentelemetry.io/otel/trace"
)

type alphaV1Ops struct {
	supervisor *suture.Supervisor
	logging    *logdrain.ServiceConfig
}

func (a *alphaV1Ops) Create(ctx context.Context, which resources.Meta, spec exec.InvocationAlphaV1Spec, bridge *operator.KindBridgeState) (*proc, exec.InvocationAlphaV1Status, error) {
	p := &proc{
		cmd:             spec.Exec,
		wd:              spec.WorkingDir,
		which:           which,
		onChange:        bridge,
		logging:         a.logging,
		startingLink:    trace.SpanContextFromContext(ctx),
		supervisionTree: a.supervisor,
	}
	a.supervisor.Add(p)
	return p, p.toAlphaV1Status(), nil
}

func (a *alphaV1Ops) Update(ctx context.Context, which resources.Meta, rt *proc, s exec.InvocationAlphaV1Spec) (exec.InvocationAlphaV1Status, error) {
	return rt.toAlphaV1Status(), nil
}
