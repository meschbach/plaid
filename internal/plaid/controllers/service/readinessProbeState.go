package service

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/controllers/tooling"
	"github.com/meschbach/plaid/internal/plaid/controllers/probes"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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

func (r *readinessProbeState) reconcile(parent context.Context, env tooling.Env, spec *probes.TemplateAlpha1Spec, status *Alpha1Status) (bool, error) {
	if spec == nil { //no readiness probes
		status.Ready = true
		return true, nil
	}

	ctx, span := tracer.Start(parent, "service.readinessProbe")
	defer span.End()

	status.Ready = false
	step, err := r.decideNextStep(ctx, env)
	if err != nil {
		span.SetStatus(codes.Error, "failed to decide what to do")
		span.RecordError(err)
		return false, err
	}
	switch step {
	case readinessProbeCreate:
		span.AddEvent("creating-probe")
		r.state, err = spec.Instantiate(ctx, probes.TemplateEnv{
			ClaimedBy: env.Subject,
			Storage:   env.Storage,
			Watcher:   env.Watcher,
			OnChange:  env.Reconcile,
		})
		if err == nil {
			r.created = true
		} else {
			span.SetStatus(codes.Error, "failed to create probe")
		}
		return false, err
	case readinessProbeWait:
		status.Ready = r.state.Ready()
		span.AddEvent("readiness-wait", trace.WithAttributes(attribute.Bool("ready", status.Ready)))
		return true, nil
	default:
		return false, fmt.Errorf("unknown state: %d", step)
	}
}

func (r *readinessProbeState) decideNextStep(ctx context.Context, env tooling.Env) (readinessProbeStep, error) {
	if !r.created {
		return readinessProbeCreate, nil
	}

	err := r.state.Reconcile(ctx, env.Storage)
	return readinessProbeWait, err
}
