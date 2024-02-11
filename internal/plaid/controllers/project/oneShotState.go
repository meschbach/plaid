package project

import (
	"context"
	"errors"
	"fmt"
	"github.com/meschbach/plaid/internal/plaid/controllers/buildrun"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/resources"
)

type resourceEnv struct {
	which     resources.Meta
	rpc       *resources.Client
	watcher   *resources.ClientWatcher
	reconcile func(ctx context.Context) error
}

type oneShotNext int

const (
	oneShotWait oneShotNext = iota
	oneShotCreate
	oneShotFinished
)

const (
	oneShotUnfinished = iota
	oneShotSuccess
	oneShotFailure
)

type oneShotState struct {
	created     bool
	ref         resources.Meta
	watchToken  resources.WatchToken
	finishState int
}

func (o *oneShotState) toStatus(status *Alpha1OneShotStatus) {
	if !o.created {
		status.State = Alpha1StateProgressing
		status.Ref = nil
		status.Done = false
	}
	status.Ref = &o.ref
	switch o.finishState {
	case oneShotSuccess:
		status.State = Alpha1StateSuccess
	case oneShotFailure:
		status.State = Alpha1StateFailed
	case oneShotUnfinished:
		status.State = Alpha1StateProgressing
	default:
		panic(fmt.Sprintf("Unhanled finish state %d", o.finishState))
	}
}

func (o *oneShotState) decideNextStep(ctx context.Context, resEnv *resourceEnv) (oneShotNext, error) {
	if !o.created {
		return oneShotCreate, nil
	}

	var procState buildrun.AlphaStatus1
	if exists, err := resEnv.rpc.GetStatus(ctx, o.ref, &procState); err != nil || !exists {
		return oneShotWait, err
	}
	if procState.Run.Result == nil || procState.Run.Result.Finished == nil {
		return oneShotWait, nil
	}

	if (*procState.Run.Result.ExitStatus) >= 0 {
		o.finishState = oneShotSuccess
	} else {
		o.finishState = oneShotFailure
	}
	return oneShotFinished, nil
}

func (o *oneShotState) create(ctx context.Context, resEnv *resourceEnv, spec Alpha1Spec, oneShotSpec Alpha1OneShotSpec) error {
	which := resEnv.which

	ref := resources.Meta{
		Type: buildrun.Alpha1,
		Name: which.Name + oneShotSpec.Name + "-" + resources.GenSuffix(4),
	}
	token, err := resEnv.watcher.OnResourceStatusChanged(ctx, ref, func(ctx context.Context, changed resources.ResourceChanged) error {
		switch changed.Operation {
		case resources.StatusUpdated:
			return resEnv.reconcile(ctx)
		default:
			return nil
		}
	})
	buildSpec := buildrun.AlphaSpec1{
		RestartToken: "", //todo: figure out restart tokens.
		Build: exec.TemplateAlpha1Spec{
			Command:    oneShotSpec.Build.Command,
			WorkingDir: spec.BaseDirectory,
		},
		Run: exec.TemplateAlpha1Spec{
			Command:    oneShotSpec.Run.Command,
			WorkingDir: spec.BaseDirectory,
		},
	}
	buildSpec.Requires = oneShotSpec.Requires

	err = resEnv.rpc.Create(ctx, ref, buildSpec, resources.ClaimedBy(which))

	if err == nil {
		o.created = true
		o.ref = ref
		o.watchToken = token
		return nil
	} else {
		unwatch := resEnv.watcher.Off(ctx, token)
		return errors.Join(err, unwatch)
	}
}

func (o *oneShotState) delete(ctx context.Context, resEnv *resourceEnv) error {
	if !o.created {
		return nil
	}
	watchError := resEnv.watcher.Off(ctx, o.watchToken)
	_, deleteErr := resEnv.rpc.Delete(ctx, o.ref)
	return errors.Join(watchError, deleteErr)
}
