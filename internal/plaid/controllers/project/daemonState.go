package project

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/controllers/service/alpha2"
	"github.com/meschbach/plaid/controllers/tooling"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/resources"
)

type daemonNext int

const (
	daemonWait daemonNext = iota
	daemonCreate
	daemonFinished
	daemonUpdate
)

func (d daemonNext) String() string {
	switch d {
	case daemonWait:
		return "daemon-wait"
	case daemonCreate:
		return "daemon-create"
	case daemonFinished:
		return "daemon-finished"
	case daemonUpdate:
		return "daemon-update"
	default:
		return fmt.Sprintf("unknown daemonNext %d", d)
	}
}

type daemonState struct {
	service     tooling.Subresource[alpha2.Status]
	targetReady bool
}

func (d *daemonState) toStatus(spec Alpha1DaemonSpec, status *Alpha1DaemonStatus) {
	status.Name = spec.Name
	status.Ready = d.targetReady
	if d.service.Created {
		status.Current = &d.service.Ref
	}
}

func (d *daemonState) decideNextStep(ctx context.Context, env tooling.Env, restartToken string) (daemonNext, error) {
	var procState alpha2.Status
	step, err := d.service.Decide(ctx, env, &procState)
	if err != nil {
		return daemonWait, err
	}
	switch step {
	case tooling.SubresourceCreate:
		return daemonCreate, nil
	case tooling.SubresourceExists:
		//todo: alpha2 should export readiness
		d.targetReady = procState.Ready
		if procState.LatestToken != restartToken {
			return daemonUpdate, nil
		}
	}
	return daemonWait, nil
}

func (d *daemonState) create(ctx context.Context, env tooling.Env, spec Alpha1Spec, daemonSpec Alpha1DaemonSpec) error {
	which := env.Subject

	resSpec := alpha2.Spec{
		Dependencies: make([]resources.Meta, len(daemonSpec.Requires)),
		Run: exec.TemplateAlpha1Spec{
			Command:    daemonSpec.Run.Command,
			WorkingDir: spec.BaseDirectory,
		},
		Readiness:    daemonSpec.Readiness,
		RestartToken: spec.RestartToken,
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
		Type: alpha2.Type,
		Name: which.Name + daemonSpec.Name + "-" + resources.GenSuffix(4),
	}
	return d.service.Create(ctx, env, ref, resSpec)
}

func (d *daemonState) update(ctx context.Context, env tooling.Env, spec Alpha1Spec, daemonSpec Alpha1DaemonSpec, restartToken string) error {
	//todo: probably should just update all of it
	var serviceSpec alpha2.Spec
	exists, err := env.Storage.Get(ctx, d.service.Ref, &serviceSpec)
	if err != nil {
		return err
	}
	if !exists {
		return d.create(ctx, env, spec, daemonSpec)
	}
	serviceSpec.RestartToken = restartToken
	exists, err = env.Storage.Update(ctx, d.service.Ref, serviceSpec)
	if err != nil {
		return err
	}
	if !exists {
		return d.create(ctx, env, spec, daemonSpec)
	}
	return nil
}

func (d *daemonState) delete(ctx context.Context, env tooling.Env) error {
	return d.service.Delete(ctx, env)
}
