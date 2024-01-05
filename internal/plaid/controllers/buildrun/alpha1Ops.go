package buildrun

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/internal/plaid/controllers/dependencies"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/operator"
)

type alpha1Ops struct {
	client  *resources.Client
	watcher *resources.ClientWatcher
}

func (a *alpha1Ops) Create(ctx context.Context, which resources.Meta, spec AlphaSpec1, bridge *operator.KindBridgeState) (*state, AlphaStatus1, error) {
	s := &state{
		bridge:   bridge,
		requires: dependencies.State{},
	}
	s.requires.Init(spec.Requires)
	status, err := a.Update(ctx, which, s, spec)
	return s, status, err
}

func (a *alpha1Ops) Update(parent context.Context, which resources.Meta, rt *state, spec AlphaSpec1) (AlphaStatus1, error) {
	if !which.Type.Equals(Alpha1) {
		panic(fmt.Sprintf("wrong type: %s\n", which.Type))
	}
	ctx, span := tracer.Start(parent, "buildrun.Update")
	defer span.End()
	status := AlphaStatus1{}

	currentRestartToken := spec.RestartToken

	runtimeEnv := &stateEnv{
		object:  which,
		rpc:     a.client,
		watcher: a.watcher,
		reconcile: func(ctx context.Context) error {
			return rt.bridge.OnResourceChange(ctx, which)
		},
		restartToken: currentRestartToken,
	}

	// Are dependencies ready?
	depsReady, depsStatus, err := rt.requires.Reconcile(ctx, dependencies.Env{
		Storage:  runtimeEnv.rpc,
		Watcher:  runtimeEnv.watcher,
		OnChange: runtimeEnv.reconcile,
	})
	status.Dependencies = depsStatus
	if err != nil || !depsReady {
		return status, err
	}

	status.Build.Token = rt.builder.lastRestartToken //todo: maybe a better way to do this
	if step, buildStatus, err := rt.builder.decideNextStep(ctx, a.client); err != nil {
		return status, err
	} else {
		status.Build = buildStatus
		switch step {
		case builderNextWait:
			return status, nil
		case builderNextCreate:
			err := rt.builder.create(ctx, runtimeEnv, spec.Build)
			if err == nil {
				status.Build.State = "created"
				status.Build.Token = rt.builder.lastRestartToken
			}
			return status, err
		default:
			return status, fmt.Errorf("unexpected builder next step: %s", step)
		case builderStateSuccessfullyCompleted: //continue on
		}
	}

	status.Run.Token = rt.proc.lastRestartToken
	if step, runStatus, err := rt.proc.decideNextStep(ctx, runtimeEnv); err != nil {
		return status, err
	} else {
		status.Run = runStatus
		switch step {
		case runStateNextWait:
			return status, nil
		case runStateNextCreate:
			err := rt.proc.create(ctx, runtimeEnv, spec.Run)
			if err == nil {
				status.Run.State = "created"
				status.Run.Token = rt.proc.lastRestartToken
			}
			return status, err
		default:
			return status, fmt.Errorf("unexpected proc next step: %s", step)
		case runStateNextFinishedSuccessfully: //continue on
		}
	}
	return status, nil
}
