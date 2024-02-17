package local

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/internal/plaid/controllers/logdrain"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/operator"
	"github.com/thejerf/suture/v4"
	"go.opentelemetry.io/otel/trace"
)

type alphaV1Ops struct {
	supervisor *suture.Supervisor
	logging    *logdrain.ServiceConfig
}

func (a *alphaV1Ops) Create(ctx context.Context, which resources.Meta, spec exec.InvocationAlphaV1Spec, bridge *operator.KindBridgeState) (*proc, exec.InvocationAlphaV1Status, error) {
	group := suture.NewSimple(which.String())
	p := &proc{
		cmd:             spec.Exec,
		wd:              spec.WorkingDir,
		which:           which,
		onChange:        bridge,
		logging:         a.logging,
		startingLink:    trace.SpanContextFromContext(ctx),
		supervisionTree: group,
	}
	group.Add(p)
	p.serviceToken = a.supervisor.Add(group)
	return p, p.toAlphaV1Status(), nil
}

func (a *alphaV1Ops) Update(ctx context.Context, which resources.Meta, rt *proc, s exec.InvocationAlphaV1Spec) (exec.InvocationAlphaV1Status, error) {
	return rt.toAlphaV1Status(), nil
}

func (a *alphaV1Ops) Delete(ctx context.Context, which resources.Meta, rt *proc) error {
	return a.supervisor.Remove(rt.serviceToken)
}
