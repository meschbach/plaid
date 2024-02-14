package project

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/internal/plaid/controllers/buildrun"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/resources"
)

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
	buildRun    Subresource[buildrun.AlphaStatus1]
	finishState int
}

func (o *oneShotState) toStatus(status *Alpha1OneShotStatus) {
	if !o.buildRun.Created {
		status.State = Alpha1StateProgressing
		status.Ref = nil
		status.Done = false
	}
	status.Ref = &o.buildRun.Ref
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
	var procState buildrun.AlphaStatus1
	step, err := o.buildRun.Decide(ctx, resEnv, &procState)
	if err != nil {
		return oneShotWait, err
	}
	switch step {
	case SubresourceCreated:
		return oneShotCreate, nil
	case SubresourceExists:
		//fall through
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

	return o.buildRun.Create(ctx, resEnv, ref, buildSpec)
}

func (o *oneShotState) delete(ctx context.Context, resEnv *resourceEnv) error {
	return o.buildRun.Delete(ctx, resEnv)
}
