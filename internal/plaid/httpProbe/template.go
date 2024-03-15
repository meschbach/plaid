package httpProbe

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/controllers/tooling"
	"github.com/meschbach/plaid/resources"
)

type TemplateAlpha1 struct {
	Port uint16 `json:"port"`
	Path string `json:"path"`
}

func (t *TemplateAlpha1) Instantiate(ctx context.Context, storage resources.Storage, claimedBy resources.Meta, watch resources.Watcher, onChange func(ctx context.Context) error) (*TemplateState, error) {
	state := &TemplateState{
		env: tooling.Env{
			Subject:   claimedBy,
			Storage:   storage,
			Watcher:   watch,
			Reconcile: onChange,
		},
		spec:  t,
		probe: tooling.Subresource[AlphaV1Status]{},
		ready: false,
	}
	return state, state.Reconcile(ctx, nil)
}

type TemplateState struct {
	env   tooling.Env
	spec  *TemplateAlpha1
	probe tooling.Subresource[AlphaV1Status]
	ready bool
}

func (t *TemplateState) Reconcile(ctx context.Context, storage resources.Storage) error {
	var status AlphaV1Status
	resourceStep, err := t.probe.Decide(ctx, t.env, &status)
	if err != nil {
		return err
	}
	switch resourceStep {
	case tooling.SubresourceCreate:
		ref := resources.Meta{
			Type: AlphaV1Type,
			Name: fmt.Sprintf("%s-port-%d", t.env.Subject.Name, t.spec.Port),
		}
		spec := AlphaV1Spec{
			Enabled:  true,
			Host:     "localhost",
			Port:     t.spec.Port,
			Resource: t.spec.Path,
		}
		if err := t.probe.Create(ctx, t.env, ref, spec, resources.ClaimedBy(t.env.Subject)); err != nil {
			return err
		}
		return nil
	case tooling.SubresourceExists:
		t.ready = status.Ready
	}
	return nil
}

func (t *TemplateState) Ready() bool {
	return t.ready
}

type TemplateAlpha1Status struct {
	Ref   *resources.Meta `json:"ref"`
	Ready bool
}

func (t *TemplateState) Status() TemplateAlpha1Status {
	if !t.probe.Created {
		return TemplateAlpha1Status{
			Ref:   nil,
			Ready: false,
		}
	}
	return TemplateAlpha1Status{
		Ref:   &t.probe.Ref,
		Ready: t.ready,
	}
}
