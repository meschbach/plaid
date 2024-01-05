package service

import (
	"context"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources/operator"
)

type dependencyNextStep int

const (
	dependencyWait dependencyNextStep = iota
	dependencySetup
	dependencyReady
)

type dependencyState struct {
	ref        resources.Meta
	setupWatch bool
	token      resources.WatchToken
}

func (d *dependencyState) decideNextStep(ctx context.Context, env resEnv) (dependencyNextStep, error) {
	if !d.setupWatch {
		return dependencySetup, nil
	}

	var status operator.ReadyStatus
	if exists, err := env.rpc.GetStatus(ctx, d.ref, &status); err != nil || !exists {
		return dependencyWait, err
	}

	if !status.Ready {
		return dependencyWait, nil
	}
	return dependencyReady, nil
}

func (d *dependencyState) setup(ctx context.Context, env resEnv) error {
	token, err := env.watcher.OnResource(ctx, d.ref, func(ctx context.Context, changed resources.ResourceChanged) error {
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
	d.token = token
	d.setupWatch = true
	return nil
}
