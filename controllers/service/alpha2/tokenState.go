package alpha2

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/controllers/tooling"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"time"
)

type tokenState struct {
	token        string
	spec         exec.TemplateAlpha1Spec
	run          tooling.Subresource[exec.InvocationAlphaV1Status]
	lastModified time.Time
}

func (t *tokenState) progressBuild(ctx context.Context, env tooling.Env, s *State) error {
	var status exec.InvocationAlphaV1Status
	step, err := t.run.Decide(ctx, env, &status)
	if err != nil {
		return err
	}
	switch step {
	case tooling.SubresourceCreate:
		t.lastModified = time.Now()
		invocationRef, invocationSpec, err := t.spec.AsSpec(env.Subject.Name + "-" + t.token)
		if err != nil {
			return err
		}
		if err := t.run.Create(ctx, env, invocationRef, invocationSpec); err != nil {
			return err
		}
	case tooling.SubresourceExists:
		if status.Started == nil {
			return nil
		}
		if !status.Healthy {
			return nil
		}
		t.lastModified = time.Now()
		s.promoteNext()
	}
	return nil
}

func (t *tokenState) progressRun(ctx context.Context, env tooling.Env, s *State) error {
	var status exec.InvocationAlphaV1Status
	step, err := t.run.Decide(ctx, env, &status)
	if err != nil {
		return err
	}
	switch step {
	case tooling.SubresourceCreate:
		//gone missing? recreate.
		t.lastModified = time.Now()
		invocationRef, invocationSpec, err := t.spec.AsSpec(env.Subject.Name + "-")
		if err != nil {
			return err
		}
		if err := t.run.Create(ctx, env, invocationRef, invocationSpec); err != nil {
			return err
		}
	case tooling.SubresourceExists:
		//todo: ensure process is healthy
	}
	return nil
}

func (t *tokenState) progressStopping(ctx context.Context, env tooling.Env) (bool, error) {
	var status exec.InvocationAlphaV1Status
	step, err := t.run.Decide(ctx, env, &status)
	if err != nil {
		return false, err
	}
	switch step {
	case tooling.SubresourceCreate:
		return true, nil
		//does not exist, progress to old
	case tooling.SubresourceExists:
		if err := t.run.Delete(ctx, env); err != nil {
			return false, err
		}
		//done, wait until delete confirmation to move to old
		return false, nil
	default:
		return false, fmt.Errorf("unexpected subresource step: %#v\n", step)
	}
}

func (t *tokenState) toStatus() TokenStatus {
	out := TokenStatus{
		Token: t.token,
		Last:  t.lastModified,
	}
	if !t.run.Created {
		out.Stage = TokenStageInit
		return out
	}
	out.Stage = TokenStageStarting
	out.Service = &t.run.Ref
	return out
}

func (t *tokenState) toOldStatus() TokenStatus {
	return TokenStatus{
		Token:   t.token,
		Stage:   TokenStageStopped,
		Last:    t.lastModified,
		Service: &t.run.Ref,
	}
}

func (t *tokenState) toStoppingStatus() TokenStatus {
	return TokenStatus{
		Token:   t.token,
		Stage:   TokenStageStopped,
		Last:    t.lastModified,
		Service: &t.run.Ref,
	}
}
