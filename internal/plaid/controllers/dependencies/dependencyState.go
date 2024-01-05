package dependencies

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/resources"
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

func (d *DependencyState) Reconcile(ctx context.Context, env Env) (ready bool, err error) {
	step, err := d.decideNextStep(ctx, env)
	if err != nil {
		return false, err
	}
	switch step {
	case nextStepReady:
		return true, nil
	case nextStepWait:
		return false, nil
	case nextStepWatch:
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
		switch changed.Operation {
		case resources.StatusUpdated:
			return env.OnChange(ctx)
		default:
			return nil
		}
	})
	if problem != nil {
		return problem
	}
	d.watching = true
	d.watchToken = token
	return env.OnChange(ctx)
}
