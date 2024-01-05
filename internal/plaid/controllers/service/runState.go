package service

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/resources"
)

type runStateStep uint8

const (
	runStateWait runStateStep = iota
	runStateCreate
)

type runState struct {
	created        bool
	lastRun        resources.Meta
	lastWatchToken resources.WatchToken
}

func (r *runState) decideNextStep(ctx context.Context, env resEnv) (runStateStep, error) {
	if !r.created {
		return runStateCreate, nil
	}
	return runStateWait, nil
}

func (r *runState) create(ctx context.Context, env resEnv, spec exec.TemplateAlpha1Spec) error {
	res, token, err := spec.CreateResource(ctx, env.rpc, env.object, nil, env.watcher, func(ctx context.Context, changed resources.ResourceChanged) error {
		switch changed.Operation {
		case resources.StatusUpdated:
			return env.reconcile(ctx)
		default:
			return nil
		}
	})
	if err != nil {
		return err
	}
	r.created = true
	r.lastRun = res
	r.lastWatchToken = token
	return nil
}
