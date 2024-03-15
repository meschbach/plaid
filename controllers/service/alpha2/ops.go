package alpha2

import (
	"context"
	"errors"
	"github.com/meschbach/plaid/controllers/tooling"
	"github.com/meschbach/plaid/controllers/tooling/kit"
	"github.com/meschbach/plaid/internal/plaid/controllers/dependencies"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/resources"
	"time"
)

type Ops struct {
	storage  resources.Storage
	observer resources.Watcher
}

func (o *Ops) Create(ctx context.Context, which resources.Meta, spec Spec, bridge kit.Manager) (*State, error) {
	depState := dependencies.State{}
	depState.Init(spec.Dependencies)
	next := &tokenState{
		token:      spec.RestartToken,
		spec:       spec.Run,
		depState:   depState,
		depsFuse:   false,
		probesSpec: spec.Readiness,
	}
	state := &State{
		next:   next,
		bridge: bridge,
	}
	env := tooling.Env{
		Subject:   which,
		Storage:   o.storage,
		Watcher:   o.observer,
		Reconcile: bridge.UpdateState,
	}
	err := state.progress(ctx, env)
	return state, err
}

func (o *Ops) Update(ctx context.Context, which resources.Meta, rt *State, s Spec) error {
	if rt.next != nil {
		if rt.next.token != s.RestartToken {
			rt.stopping = append(rt.stopping, rt.next)
			//todo: dry
			rt.next = &tokenState{
				token:        s.RestartToken,
				spec:         s.Run,
				run:          tooling.Subresource[exec.InvocationAlphaV1Status]{},
				lastModified: time.Now(),
			}
		}
	} else if rt.stable != nil {
		if rt.stable.token != s.RestartToken {
			//kick off new build
			rt.next = &tokenState{
				token:        s.RestartToken,
				spec:         s.Run,
				run:          tooling.Subresource[exec.InvocationAlphaV1Status]{},
				lastModified: time.Now(),
			}
		}
	} else {
		return errors.New("unexpected -- no build or stable")
	}
	//todo: kit should probably call this for us.
	return o.UpdateState(ctx, which, rt)
}

func (o *Ops) UpdateState(ctx context.Context, which resources.Meta, rt *State) error {
	env := tooling.Env{
		Subject:   which,
		Storage:   o.storage,
		Watcher:   o.observer,
		Reconcile: rt.bridge.UpdateState,
	}
	return rt.progress(ctx, env)
}

func (o *Ops) Delete(ctx context.Context, which resources.Meta, rt *State) error {
	env := tooling.Env{
		Subject:   which,
		Storage:   o.storage,
		Watcher:   o.observer,
		Reconcile: rt.bridge.UpdateState,
	}

	var problems []error
	if rt.next != nil {
		problems = append(problems, rt.next.delete(ctx, env))
		rt.next = nil
	}
	if rt.stable != nil {
		problems = append(problems, rt.stable.delete(ctx, env))
		rt.stable = nil
	}
	for _, stopping := range rt.stopping {
		problems = append(problems, stopping.delete(ctx, env))
	}
	rt.stopping = nil
	for _, old := range rt.old {
		problems = append(problems, old.delete(ctx, env))
	}
	rt.old = nil
	return errors.Join(problems...)
}

func (o *Ops) Status(ctx context.Context, rt *State) Status {
	status := Status{}
	if rt.stable != nil {
		stableStatus := rt.stable.toStatus()
		status.LatestToken = stableStatus.Token
		status.Stable = &stableStatus
		status.Ready = stableStatus.Ready
	}
	if rt.next != nil {
		nextStatus := rt.next.toStatus()
		status.LatestToken = nextStatus.Token
		status.Next = &nextStatus
		status.Ready = false
	}
	status.Old = make([]TokenStatus, 0, len(rt.old)+len(rt.stopping))
	for _, old := range rt.old {
		status.Old = append(status.Old, old.toOldStatus())
	}
	for _, stopping := range rt.stopping {
		status.Old = append(status.Old, stopping.toStoppingStatus())
	}
	return status
}
