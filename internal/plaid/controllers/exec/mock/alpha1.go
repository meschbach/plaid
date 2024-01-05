package mock

import (
	"context"
	"errors"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/internal/plaid/controllers/logdrain"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/operator"
)

type alpha1Ops struct {
	push   *resources.Client
	engine *ExecEngine
}

func (a *alpha1Ops) Create(ctx context.Context, which resources.Meta, spec exec.InvocationAlphaV1Spec, bridge *operator.KindBridgeState) (*Proc, exec.InvocationAlphaV1Status, error) {
	stdout := streams.NewBuffer[logdrain.LogEntry](32)
	runtimeEntity := &Proc{
		interpreter: a,
		object:      which,
		StdOut:      stdout,
	}
	//todo: consider when this should actually happen
	a.engine.reactor.ScheduleStateFunc(ctx, func(ctx context.Context, state *engineState) error {
		state.procs.Upsert(which, runtimeEntity)
		a.engine.logging.RegisterSource(ctx, which, "stdout", stdout)
		return a.engine.events.Emit(ctx, state)
	})
	return runtimeEntity, exec.InvocationAlphaV1Status{}, nil
}

func (a *alpha1Ops) Update(ctx context.Context, which resources.Meta, rt *Proc, s exec.InvocationAlphaV1Spec) (exec.InvocationAlphaV1Status, error) {
	return exec.InvocationAlphaV1Status{}, errors.New("updated todo")
}

func (a *alpha1Ops) updateProc(ctx context.Context, m *Proc) error {
	exit := &m.exitCode
	status := exec.InvocationAlphaV1Status{
		Started:    m.startedAt,
		Finished:   m.finishedAt,
		ExitStatus: exit,
		Healthy:    false,
	}
	_, err := a.push.UpdateStatus(ctx, m.object, status)
	if err != nil {
		return err
	}
	return nil
}
