package buildrun

import (
	"context"
	"fmt"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/controllers/exec"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
)

type runStateNext int

const (
	runStateNextWait runStateNext = iota
	runStateNextCreate
	runStateNextFinishedSuccessfully
)

func (r runStateNext) String() string {
	switch r {
	case runStateNextWait:
		return "wait"
	case runStateNextCreate:
		return "create"
	case runStateNextFinishedSuccessfully:
		return "success"
	default:
		return fmt.Sprintf("unknown run state next %d", r)
	}
}

type runState struct {
	cratedRun        bool
	lastRun          resources.Meta
	lastWatchToken   resources.WatchToken
	lastRestartToken string
}

func (r *runState) decideNextStep(ctx context.Context, env *stateEnv) (runStateNext, Alpha1StatusRun, error) {
	status := Alpha1StatusRun{
		Token: r.lastRestartToken,
		State: "controller-error",
		Ref:   &r.lastRun,
	}
	if !r.cratedRun {
		return runStateNextCreate, status, nil
	}

	var procStatus exec.InvocationAlphaV1Status
	exists, err := env.rpc.GetStatus(ctx, r.lastRun, &procStatus)
	if err != nil {
		return runStateNextWait, status, err
	}
	if !exists {
		status.State = "creating"
		return runStateNextWait, status, err
	}
	status.Result = &procStatus

	if procStatus.Finished == nil {
		status.State = "running"
		return runStateNextWait, status, nil
	}
	status.State = "finished"
	return runStateNextFinishedSuccessfully, status, nil
}

func (r *runState) create(ctx context.Context, env *stateEnv, spec exec.TemplateAlpha1Spec) error {
	annotations := make(map[string]string)
	annotations[procAnnotationRole] = procAnnotationRoleProc
	annotations[procAnnotationToken] = env.restartToken

	ref, watchToken, err := spec.CreateResource(ctx, env.rpc, env.object, annotations, env.watcher, func(ctx context.Context, changed resources.ResourceChanged) error {
		switch changed.Operation {
		case resources.StatusUpdated:
			return env.reconcile(ctx)
		}
		return nil
	})
	if err != nil {
		return err
	}
	r.cratedRun = true
	r.lastRun = ref
	r.lastWatchToken = watchToken
	r.lastRestartToken = env.restartToken
	return nil
}
