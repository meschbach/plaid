package service

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/internal/plaid/controllers/probes"
)

type readinessProbeStep uint8

const (
	readinessProbeCreate readinessProbeStep = iota
	//no work to be done, continue to wait
	readinessProbeWait
)

type readinessProbeState struct {
	created bool
	state   probes.TemplateAlpha1State
}

func (r *readinessProbeState) reconcile(ctx context.Context, env *resEnv, spec *probes.TemplateAlpha1Spec, status *Alpha1Status) (bool, error) {
	step, err := r.decideNextStep(ctx, env)
	if err != nil {
		return false, err
	}
	switch step {
	case readinessProbeCreate:
		r.state, err = spec.Instantiate(ctx, probes.TemplateEnv{
			ClaimedBy: env.object,
			Storage:   env.rpc,
			Watcher:   env.watcher,
			OnChange:  env.reconcile,
		})
		if err == nil {
			r.created = true
		}
		return false, err
	case readinessProbeWait:
		status.Ready = r.state.Ready()
		return true, nil
	default:
		return false, fmt.Errorf("unknown state: %d", step)
	}
}

func (r *readinessProbeState) decideNextStep(ctx context.Context, env *resEnv) (readinessProbeStep, error) {
	if !r.created {
		return readinessProbeCreate, nil
	}

	err := r.state.Reconcile(ctx, env.rpc)
	return readinessProbeWait, err
}
