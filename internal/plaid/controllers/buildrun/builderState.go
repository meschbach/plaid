package buildrun

import (
	"context"
	"fmt"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/controllers/exec"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
)

type builderState struct {
	createdBuild     bool
	lastBuild        resources.Meta
	lastWatchToken   resources.WatchToken
	lastRestartToken string
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

func (b *builderState) decideNextStep(ctx context.Context, c *resources.Client) (builderNextStep, Alpha1StatusBuild, error) {
	status := Alpha1StatusBuild{
		Token: b.lastRestartToken,
		State: "controller-error",
		Ref:   &b.lastBuild,
	}
	if !b.createdBuild {
		status.State = "creating"
		return builderNextCreate, status, nil
	}

	var procStatus exec.InvocationAlphaV1Status
	exists, err := c.GetStatus(ctx, b.lastBuild, &procStatus)
	if err != nil {
		return builderNextWait, status, err
	}
	if !exists {
		status.State = "proc-creating"
		return builderNextWait, status, nil
	} //need to wait for its existence
	status.State = "proc-exec"
	status.Result = &procStatus
	status.Token = b.lastRestartToken

	if procStatus.Finished == nil {
		return builderNextWait, status, nil
	}
	status.State = "completed"
	return builderStateSuccessfullyCompleted, status, nil
}

func (b *builderState) create(ctx context.Context, env *stateEnv, spec exec.TemplateAlpha1Spec) error {
	annotations := make(map[string]string)
	annotations[procAnnotationRole] = procAnnotationRoleBuilder
	annotations[procAnnotationToken] = env.restartToken
	//todo: annotation to reacquire in case of failure
	ref, watchToken, err := spec.CreateResource(ctx, env.rpc, env.object, annotations, env.watcher, func(ctx context.Context, changed resources.ResourceChanged) error {
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
	b.lastRestartToken = env.restartToken
	return nil
}
