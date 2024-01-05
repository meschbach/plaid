package projectfile

import (
	"context"
	"errors"
	"fmt"
	"github.com/meschbach/plaid/internal/plaid/controllers/dependencies"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/internal/plaid/controllers/project"
	"github.com/meschbach/plaid/resources"
)

type projectNext int

func (p projectNext) String() string {
	switch p {
	case projectNextWait:
		return "wait"
	case projectNextCreate:
		return "create"
	case projectNextDone:
		return "done"
	default:
		return fmt.Sprintf("unknown(%d)", p)
	}
}

const (
	projectNextWait projectNext = iota
	projectNextCreate
	projectNextDone
)

type projectState struct {
	created       bool
	finished      bool
	finishedState string
	current       resources.Meta
	currentToken  resources.WatchToken
}

func (p *projectState) updateStatus(ctx context.Context, status *Alpha1Status) {
	if p.created {
		status.Current = &p.current
		status.Done = p.finished
		if p.finished {
			status.Success = p.finishedState == "success"
		}
	}
	//todo: test when not created
}

func (p *projectState) decideNextSteps(ctx context.Context, env stateEnv) (projectNext, error) {
	if !p.created {
		return projectNextCreate, nil
	}

	var projectStatus project.Alpha1Status
	if exists, err := env.rpc.GetStatus(ctx, p.current, &projectStatus); err != nil {
		return projectNextWait, err
	} else if !exists {
		return projectNextWait, nil
	}

	if projectStatus.Done {
		p.finished = true
		if projectStatus.Result == project.Alpha1StateSuccess {
			p.finishedState = "success"
		} else {
			p.finishedState = "failed"
		}
		return projectNextDone, nil
	}
	p.finished = false
	return projectNextWait, nil
}

func (p *projectState) create(ctx context.Context, env stateEnv, spec Alpha1Spec, config Configuration) error {
	p.created = true
	subspec := project.Alpha1Spec{
		BaseDirectory: spec.WorkingDirectory,
	}
	if config.IsOneShot() {
		oneShotSpec := project.Alpha1OneShotSpec{
			Name: config.Name,
			Run: exec.TemplateAlpha1Spec{
				Command:    config.Run,
				WorkingDir: spec.WorkingDirectory,
			},
		}
		if config.Build != nil {
			oneShotSpec.Build = exec.TemplateAlpha1Spec{
				Command:    config.Build.Exec,
				WorkingDir: spec.WorkingDirectory,
			}
		}
		for _, dep := range config.Requires {
			oneShotSpec.Requires = append(oneShotSpec.Requires, dependencies.NamedDependencyAlpha1{
				Name: dep,
				Ref:  resources.Meta{Type: project.Alpha1, Name: dep},
			})
		}
		subspec.OneShots = append(subspec.OneShots, oneShotSpec)
	} else {
		d, err := config.toServiceConfig(spec.WorkingDirectory)
		if err != nil {
			return err
		}
		subspec.Daemons = append(subspec.Daemons, d)
	}

	ref := resources.Meta{
		Type: project.Alpha1,
		Name: env.which.Name,
	}
	token, err := env.watcher.OnResourceStatusChanged(ctx, ref, func(ctx context.Context, changed resources.ResourceChanged) error {
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
	p.current = ref
	p.currentToken = token

	err = env.rpc.Create(ctx, ref, subspec, resources.ClaimedBy(env.which))
	if err != nil {
		p.created = false
		unwatch := env.watcher.Off(ctx, p.currentToken)
		return errors.Join(unwatch, err)
	}
	return err
}
