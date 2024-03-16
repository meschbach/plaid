package alpha2

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/controllers/tooling"
	"github.com/meschbach/plaid/internal/plaid/controllers/dependencies"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/internal/plaid/controllers/probes"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"time"
)

type tokenState struct {
	token string
	spec  exec.TemplateAlpha1Spec
	//depsFuse is triggered when all dependencies are ready
	depState     dependencies.State
	depsFuse     bool
	depStatus    dependencies.Alpha1Status
	build        tooling.Subresource[exec.InvocationAlphaV1Status]
	buildSpec    *exec.TemplateAlpha1Spec
	buildFuse    bool
	run          tooling.Subresource[exec.InvocationAlphaV1Status]
	lastModified time.Time
	probesSpec   *probes.TemplateAlpha1Spec
	probesState  probes.TemplateAlpha1State
	probesFuse   bool
	probesReady  bool
}

func (t *tokenState) progressBuild(ctx context.Context, env tooling.Env, s *State) error {
	span := trace.SpanFromContext(ctx)
	//todo: utility for dependency state doesn't map to this system well
	if !t.depsFuse {
		ready, status, err := t.depState.Reconcile(ctx, dependencies.Env(env))
		if err != nil {
			span.SetStatus(codes.Error, "failed to update dependency")
			span.AddEvent("dependencies reconciliation failed")
			return err
		}
		t.depStatus = status
		if !ready {
			return nil
		}
		t.depsFuse = true
	}

	if t.buildSpec != nil {
		var buildStatus exec.InvocationAlphaV1Status
		buildStep, err := t.build.Decide(ctx, env, &buildStatus)
		if err != nil {
			return err
		}
		switch buildStep {
		case tooling.SubresourceCreate:
			t.lastModified = time.Now()
			buildRef, buildSpec, err := t.buildSpec.AsSpec(env.Subject.Name + "-build-" + t.token)
			if err != nil {
				return err
			}
			if err := t.build.Create(ctx, env, buildRef, buildSpec); err != nil {
				return err
			}
			return nil
		case tooling.SubresourceExists:
			if !t.buildFuse {
				//if we have not yet finished then keep going
				if buildStatus.Finished == nil || buildStatus.ExitStatus == nil {
					return nil
				}
				t.buildFuse = true
			}
			//otherwise we are good
		}
	}

	var runStatus exec.InvocationAlphaV1Status
	runStep, err := t.run.Decide(ctx, env, &runStatus)
	if err != nil {
		return err
	}
	switch runStep {
	case tooling.SubresourceCreate:
		t.lastModified = time.Now()
		invocationRef, invocationSpec, err := t.spec.AsSpec(env.Subject.Name + "-" + t.token)
		if err != nil {
			span.AddEvent("invocation spec failed")
			return err
		}
		if err := t.run.Create(ctx, env, invocationRef, invocationSpec); err != nil {
			span.AddEvent("invocation creation failed")
			return err
		}
		return nil
	case tooling.SubresourceExists:
		if runStatus.Started == nil {
			return nil
		}
		if !runStatus.Healthy {
			return nil
		}
	}
	if !t.probesFuse {
		if t.probesState == nil {
			probeEnv := probes.TemplateEnv{
				ClaimedBy: env.Subject,
				Storage:   env.Storage,
				Watcher:   env.Watcher,
				OnChange:  env.Reconcile,
			}
			probeState, err := t.probesSpec.Instantiate(ctx, probeEnv)
			if err != nil {
				return err
			}
			t.probesState = probeState
			t.lastModified = time.Now()
		}
		if err := t.probesState.Reconcile(ctx, env.Storage); err != nil {
			return err
		}
		ready := t.probesState.Ready()
		if !ready {
			return nil
		}
		t.probesFuse = true
	}
	t.probesReady = true
	t.lastModified = time.Now()
	s.promoteNext()
	return nil
}

func (t *tokenState) progressRun(ctx context.Context, env tooling.Env, s *State) error {
	//update probes state
	if err := t.probesState.Reconcile(ctx, env.Storage); err != nil {
		return err
	}
	//
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
		Token:    t.token,
		Last:     t.lastModified,
		Ready:    t.probesReady,
		Deps:     t.depStatus,
		DepsFuse: t.depsFuse,
	}
	if !t.depsFuse {
		out.Stage = TokenStageDependencyWait
		return out
	}
	if t.buildSpec != nil {
		if !t.build.Created {
			out.Stage = TokenStageBuilding
			return out
		}
		out.Build = &t.build.Ref
		if !t.buildFuse {
			return out
		}
	}
	if !t.run.Created {
		out.Stage = TokenStageInit
		return out
	}
	out.Service = &t.run.Ref
	if t.probesState == nil {
		out.Ready = false
		out.Stage = TokenStageProbeWait
		return out
	} else {
		if !t.probesFuse {
			out.Probe = t.probesState.Status()
			out.Stage = TokenStageProbeWait
			return out
		}
		out.Probe = t.probesState.Status()
		out.Ready = t.probesState.Ready()
	}
	out.Stage = TokenStageStarting
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

func (t *tokenState) delete(ctx context.Context, env tooling.Env) error {
	if t.run.Created {
		_, err := env.Storage.Delete(ctx, t.run.Ref)
		return err
	}
	return nil
}
