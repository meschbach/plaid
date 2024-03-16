package project

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/controllers/service/alpha2"
	"github.com/meschbach/plaid/controllers/tooling"
	"github.com/meschbach/plaid/internal/plaid/controllers/dependencies"
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
	service          tooling.Subresource[alpha2.Status]
	targetReady      bool
	lastRestartToken string
	nextRestartToken string
}

func (d *daemonState) toStatus(spec Alpha1DaemonSpec, status *Alpha1DaemonStatus) {
	status.Name = spec.Name
	status.Ready = d.targetReady
	if d.service.Created {
		status.Current = &d.service.Ref
	}
}

func (d *daemonState) decideNextStep(ctx context.Context, env tooling.Env) (daemonNext, error) {
	var procState alpha2.Status
	step, err := d.service.Decide(ctx, env, &procState)
	if err != nil {
		return daemonWait, err
	}
	switch step {
	case tooling.SubresourceCreate:
		return daemonCreate, nil
	case tooling.SubresourceExists:
		if d.lastRestartToken != d.nextRestartToken {
			d.targetReady = false
			return daemonUpdate, nil
		}
		//todo: alpha2 should export readiness
		d.targetReady = procState.Ready
	}
	return daemonWait, nil
}

func (d *daemonState) generateSpec(spec Alpha1Spec, daemonSpec Alpha1DaemonSpec) alpha2.Spec {
	resSpec := alpha2.Spec{
		Dependencies: make(dependencies.Alpha1Spec, len(daemonSpec.Requires)),
		Run: exec.TemplateAlpha1Spec{
			Command:    daemonSpec.Run.Command,
			WorkingDir: spec.BaseDirectory,
		},
		Readiness:    daemonSpec.Readiness,
		RestartToken: d.nextRestartToken,
	}
	if daemonSpec.Build != nil {
		resSpec.Build = &exec.TemplateAlpha1Spec{
			Command:    daemonSpec.Build.Command,
			WorkingDir: spec.BaseDirectory,
		}
	}

	for i, dep := range daemonSpec.Requires {
		resSpec.Dependencies[i] = dependencies.NamedDependencyAlpha1{
			Name: dep.Name,
			Ref:  dep,
		}
	}
	return resSpec
}

func (d *daemonState) create(ctx context.Context, env tooling.Env, spec Alpha1Spec, daemonSpec Alpha1DaemonSpec) error {
	which := env.Subject

	resSpec := d.generateSpec(spec, daemonSpec)
	ref := resources.Meta{
		Type: alpha2.Type,
		Name: which.Name + daemonSpec.Name + "-" + resources.GenSuffix(4),
	}
	return d.service.Create(ctx, env, ref, resSpec)
}

func (d *daemonState) update(ctx context.Context, env tooling.Env, spec Alpha1Spec, daemonSpec Alpha1DaemonSpec) error {
	serviceSpec := d.generateSpec(spec, daemonSpec)
	_, err := env.Storage.Update(ctx, d.service.Ref, serviceSpec)
	if err != nil {
		return err
	}
	d.lastRestartToken = d.nextRestartToken
	return nil
}

func (d *daemonState) updateRestartToken(nextToken string) {
	d.nextRestartToken = nextToken
}

func (d *daemonState) delete(ctx context.Context, env tooling.Env) error {
	return d.service.Delete(ctx, env)
}
