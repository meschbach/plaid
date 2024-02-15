package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/meschbach/plaid/controllers/tooling"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
)

// todo think about merging with buildrun.builderState as they effectively do the same thing
type builderState struct {
	buildExec tooling.Subresource[exec.InvocationAlphaV1Status]
}

type builderNextStep int

const (
	builderNextWait builderNextStep = iota
	builderNextCreate
	builderStateSuccessfullyCompleted
)

func (b builderNextStep) String() string {
	switch b {
	case builderNextWait:
		return "wait"
	case builderNextCreate:
		return "create"
	case builderStateSuccessfullyCompleted:
		return "successfully completed"
	default:
		return fmt.Sprintf("Unknown builder next state %d", b)
	}
}

func (b *builderState) decideNextStep(ctx context.Context, env tooling.Env) (builderNextStep, Alpha1BuildStatus, error) {
	status := Alpha1BuildStatus{
		State: "controller-error",
	}
	var buildCommandStatus exec.InvocationAlphaV1Status
	subresourceStep, err := b.buildExec.Decide(ctx, env, &buildCommandStatus)
	if err != nil {
		return builderNextWait, status, err
	}
	switch subresourceStep {
	case tooling.SubresourceCreate:
		status.State = "create"
		return builderNextCreate, status, err
	case tooling.SubresourceExists:
		step := builderNextWait
		status.Ref = &b.buildExec.Ref
		if buildCommandStatus.Started == nil {
			status.State = "starting"
		} else if buildCommandStatus.Finished == nil {
			status.State = "running"
		} else if buildCommandStatus.ExitStatus == nil {
			status.State = "finishing"
		} else {
			status.State = "finished"
			step = builderStateSuccessfullyCompleted
		}
		return step, status, nil
	default:
		return builderNextWait, status, errors.New("unknown subresource state " + subresourceStep.String())
	}
}

func (b *builderState) create(ctx context.Context, env tooling.Env, templateSpec *exec.TemplateAlpha1Spec, status *Alpha1BuildStatus) error {
	ref, spec, err := templateSpec.AsSpec(env.Subject.Name)
	if err != nil {
		return err
	}
	if err := b.buildExec.Create(ctx, env, ref, spec); err != nil {
		return err
	}
	status.Ref = &ref
	return nil
}

func (b *builderState) delete(ctx context.Context, env tooling.Env) error {
	return b.buildExec.Delete(ctx, env)
}
