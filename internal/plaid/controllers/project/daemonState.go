package project

import (
	"context"
	"errors"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/controllers/exec"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/controllers/service"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
)

type daemonNext int

const (
	daemonWait daemonNext = iota
	daemonCreate
	daemonFinished
)

type daemonState struct {
	created     bool
	ref         resources.Meta
	watchToken  resources.WatchToken
	targetReady bool
}

func (d *daemonState) toStatus(spec Alpha1DaemonSpec, status *Alpha1DaemonStatus) {
	status.Name = spec.Name
	if d.created {
		status.Current = &d.ref
		status.Ready = d.targetReady
	}
}

func (d *daemonState) decideNextStep(ctx context.Context, env *resourceEnv) (daemonNext, error) {
	if !d.created {
		return daemonCreate, nil
	}

	var procState service.Alpha1Status
	if exists, err := env.rpc.GetStatus(ctx, d.ref, &procState); err != nil || !exists {
		return daemonWait, err
	}
	//todo: add step for failed states
	d.targetReady = procState.Ready

	return daemonWait, nil
	//if procState.Run.Result == nil || procState.Run.Result.Finished == nil {
	//	return daemonWait, nil
	//}
	//return daemonFinished, nil
}

func (d *daemonState) create(ctx context.Context, env *resourceEnv, spec Alpha1Spec, daemonSpec Alpha1DaemonSpec) error {
	which := env.which

	resSpec := service.Alpha1Spec{
		Run: exec.TemplateAlpha1Spec{
			Command:    daemonSpec.Run.Command,
			WorkingDir: spec.BaseDirectory,
		},
		Dependencies: make([]resources.Meta, len(daemonSpec.Requires)),
		Readiness:    daemonSpec.Readiness,
	}
	if daemonSpec.Build != nil {
		resSpec.Build = &exec.TemplateAlpha1Spec{
			Command:    daemonSpec.Build.Command,
			WorkingDir: spec.BaseDirectory,
		}
	}

	for i, dep := range daemonSpec.Requires {
		resSpec.Dependencies[i] = dep
	}

	ref := resources.Meta{
		Type: service.Alpha1,
		Name: which.Name + "-daemon-" + daemonSpec.Name + "-" + resources.GenSuffix(4),
	}
	token, err := env.watcher.OnResourceStatusChanged(ctx, ref, func(ctx context.Context, changed resources.ResourceChanged) error {
		switch changed.Operation {
		case resources.StatusUpdated:
			return env.reconcile(ctx)
		default:
			return nil
		}
	})
	err = env.rpc.Create(ctx, ref, resSpec, resources.ClaimedBy(which))

	if err == nil {
		d.created = true
		d.ref = ref
		d.watchToken = token
		return nil
	} else {
		unwatch := env.watcher.Off(ctx, token)
		return errors.Join(err, unwatch)
	}
}
