package dependencies

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/resources"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// DependencyState will manage a single Dependency to monitor for readiness
type DependencyState struct {
	ref        resources.Meta
	watching   bool
	watchToken resources.WatchToken
}

func (d *DependencyState) Init(ref resources.Meta) {
	d.ref = ref
}

func (d *DependencyState) Reconcile(parent context.Context, env Env) (ready bool, err error) {
	ctx, span := tracer.Start(parent, "DependencyState.Reconcile", trace.WithAttributes(attribute.Stringer("ref", d.ref)))
	defer span.End()

	step, err := d.decideNextStep(ctx, env)
	if err != nil {
		span.SetStatus(codes.Error, "failed to decide on next step")
		span.RecordError(err)
		return false, err
	}
	switch step {
	case nextStepReady:
		span.AddEvent("dependency-ready")
		return true, nil
	case nextStepWait:
		span.AddEvent("dependency-wait")
		return false, nil
	case nextStepWatch:
		span.AddEvent("dependency-watch")
		problem := d.watch(ctx, env)
		return false, problem
	default:
		panic(fmt.Sprintf("unknown next step %d", step))
	}
}

type dependencyNext uint8

const (
	nextStepWait dependencyNext = iota
	nextStepWatch
	nextStepReady
)

func (d dependencyNext) String() string {
	switch d {
	case nextStepWait:
		return "wait"
	case nextStepWatch:
		return "watch"
	case nextStepReady:
		return "ready"
	default:
		panic(fmt.Sprintf("unkown next dependency step %d", d))
	}
}

func (d *DependencyState) decideNextStep(ctx context.Context, env Env) (dependencyNext, error) {
	if !d.watching {
		return nextStepWatch, nil
	}

	var ready ReadyAlpha1Status
	exists, err := env.Storage.GetStatus(ctx, d.ref, &ready)
	if err != nil {
		return nextStepWait, nil
	}
	if !exists {
		return nextStepWait, nil
	}

	if ready.Ready {
		return nextStepReady, nil
	} else {
		return nextStepWait, nil
	}
}

func (d *DependencyState) watch(ctx context.Context, env Env) error {
	token, problem := env.Watcher.OnResourceStatusChanged(ctx, d.ref, func(ctx context.Context, changed resources.ResourceChanged) error {
		return env.Reconcile(ctx)
	})
	if problem != nil {
		return problem
	}
	d.watching = true
	d.watchToken = token
	return env.Reconcile(ctx)
}
