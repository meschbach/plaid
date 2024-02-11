package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/resources"
)

// todo think about merging with buildrun.builderState as they effectively do the same thing
type builderState struct {
	createdBuild   bool
	lastBuild      resources.Meta
	lastWatchToken resources.WatchToken
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

func (b *builderState) decideNextStep(ctx context.Context, env resEnv) (builderNextStep, Alpha1BuildStatus, error) {
	status := Alpha1BuildStatus{
		State: "controller-error",
	}
	if !b.createdBuild {
		status.State = "creating"
		return builderNextCreate, status, nil
	}
	status.State = "created"

	var procStatus exec.InvocationAlphaV1Status
	exists, err := env.rpc.GetStatus(ctx, b.lastBuild, &procStatus)
	if err != nil {
		status.State = "internal-error"
		return builderNextWait, status, err
	}
	if !exists {
		status.State = "proc-creating"
		return builderNextWait, status, nil
	} //need to wait for its existence
	status.State = "proc-exec"

	if procStatus.Finished == nil {
		return builderNextWait, status, nil
	}
	status.State = "completed"
	return builderStateSuccessfullyCompleted, status, nil
}

func (b *builderState) create(ctx context.Context, env resEnv, spec *exec.TemplateAlpha1Spec) error {
	//todo: annotation to reacquire in case of failure
	ref, watchToken, err := spec.CreateResource(ctx, env.rpc, env.object, nil, env.watcher, func(ctx context.Context, changed resources.ResourceChanged) error {
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
	b.createdBuild = true
	b.lastBuild = ref
	b.lastWatchToken = watchToken
	return nil
}

func (b *builderState) delete(ctx context.Context, env resEnv) error {
	if !b.createdBuild {
		return nil
	}
	watchError := env.watcher.Off(ctx, b.lastWatchToken)
	_, deleteError := env.rpc.Delete(ctx, b.lastBuild)
	return errors.Join(watchError, deleteError)
}
