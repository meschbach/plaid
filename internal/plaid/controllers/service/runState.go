package service

import (
	"context"
	"errors"
	"github.com/meschbach/plaid/controllers/tooling"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/resources"
)

type runStateStep uint8

const (
	runStateWait runStateStep = iota
	runStateCreate
	runStateRunning
)

type runState struct {
	exec tooling.Subresource[exec.InvocationAlphaV1Status]
}

func (r *runState) decideNextStep(ctx context.Context, env resEnv) (runStateStep, error) {
	var status exec.InvocationAlphaV1Status
	step, err := r.exec.Decide(ctx, env.toTooling(), &status)
	if err != nil {
		return runStateWait, err
	}
	switch step {
	case tooling.SubresourceCreate:
		return runStateCreate, nil
	case tooling.SubresourceExists:
		if status.Started == nil {
			return runStateWait, nil
		} else {
			return runStateRunning, nil
		}
	default:
		return runStateWait, errors.New("bad subresource state: " + step.String())
	}
}

func (r *runState) create(ctx context.Context, env resEnv, spec exec.TemplateAlpha1Spec) error {
	ref, proposedSpec, err := spec.AsSpec(env.object.Name)
	if err != nil {
		return err
	}
	return r.exec.Create(ctx, env.toTooling(), ref, proposedSpec, resources.ClaimedBy(env.object))
}

func (r *runState) delete(ctx context.Context, env resEnv) error {
	return r.exec.Delete(ctx, env.toTooling())
}

func (r *runState) toStatus() Alpha1RunStatus {
	if !r.exec.Created {
		return Alpha1RunStatus{
			State: StateNotReady,
		}
	}
	return Alpha1RunStatus{
		State: Running,
		Ref:   &r.exec.Ref,
	}
}
