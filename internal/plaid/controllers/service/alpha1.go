package service

import (
	"context"
	"errors"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/internal/plaid/controllers/probes"
	"github.com/meschbach/plaid/internal/plaid/resources"
	"github.com/meschbach/plaid/internal/plaid/resources/operator"
)

var Alpha1 = resources.Type{
	Kind:    Kind,
	Version: "alpha1",
}

type Alpha1Spec struct {
	// Dependencies are checked for the status with a `ready` field.  If all dependencies are ready then the a build and
	// run is issued
	Dependencies []resources.Meta         `json:"dependencies,omitempty"`
	Build        *exec.TemplateAlpha1Spec `json:"build,omitempty"`
	Run          exec.TemplateAlpha1Spec  `json:"run"`
	// Readiness defines a probe to determine if the system is ready
	Readiness *probes.TemplateAlpha1Spec `json:"readiness,omitempty"`
}

type Alpha1Status struct {
	Dependencies []Alpha1StatusDependency `json:"dependencies,omitempty"`
	Build        Alpha1BuildStatus        `json:"build,omitempty"`
	Ready        bool                     `json:"ready"`
}

type Alpha1StatusDependency struct {
	Dependency resources.Meta `json:"ref"`
	Ready      bool           `json:"ready"`
}

type Alpha1BuildStatus struct {
	State string `json:"state"`
}

type alpha1Ops struct {
	client  *resources.Client
	watcher *resources.ClientWatcher
}

func (a *alpha1Ops) Create(ctx context.Context, which resources.Meta, spec Alpha1Spec, bridge *operator.KindBridgeState) (*serviceState, Alpha1Status, error) {
	rt := &serviceState{
		bridge: bridge,
	}
	status, err := a.Update(ctx, which, rt, spec)
	return rt, status, err
}

func (a *alpha1Ops) Update(ctx context.Context, which resources.Meta, rt *serviceState, s Alpha1Spec) (Alpha1Status, error) {
	env := resEnv{
		object:  which,
		rpc:     a.client,
		watcher: a.watcher,
		reconcile: func(ctx context.Context) error {
			return rt.bridge.OnResourceChange(ctx, which)
		},
	}

	status := Alpha1Status{}

	if rt.dependencies == nil {
		rt.dependencies = make([]*dependencyState, len(s.Dependencies))
	}
	status.Dependencies = make([]Alpha1StatusDependency, len(s.Dependencies))
	var dependencyErrors []error
	allReady := true
	for i, depSpec := range s.Dependencies {
		if rt.dependencies[i] == nil {
			rt.dependencies[i] = &dependencyState{
				ref: depSpec,
			}
			if err := rt.dependencies[i].setup(ctx, env); err != nil {
				allReady = false
				dependencyErrors = append(dependencyErrors, err)
				continue
			}
		}
		if next, err := rt.dependencies[i].decideNextStep(ctx, env); err != nil {
			allReady = false
			dependencyErrors = append(dependencyErrors, err)
			continue
		} else {
			switch next {
			case dependencyWait:
				status.Dependencies[i] = Alpha1StatusDependency{
					Dependency: rt.dependencies[i].ref,
					Ready:      false,
				}
				allReady = false
			case dependencyReady:
				status.Dependencies[i] = Alpha1StatusDependency{
					Dependency: rt.dependencies[i].ref,
					Ready:      true,
				}
			case dependencySetup:
				panic("must be done during creation")
			}
		}
	}
	if dependencyErrors != nil || !allReady {
		status.Ready = false
		return status, errors.Join(dependencyErrors...)
	}

	//setup build
	if s.Build != nil { //todo: test builder branch
		if step, buildStatus, err := rt.build.decideNextStep(ctx, env); err != nil {
			return status, err
		} else {
			switch step {
			case builderNextCreate:
				status.Build = buildStatus
				if err := rt.build.create(ctx, env, s.Build); err != nil {
					status.Build.State = "internal-error"
					return status, err
				}
			case builderNextWait:
				status.Build = buildStatus
				return status, err
			case builderStateSuccessfullyCompleted:
				//continue
			}
		}
	}

	//run service
	if step, err := rt.run.decideNextStep(ctx, env); err != nil {
		return status, err
	} else {
		switch step {
		case runStateCreate:
			if err := rt.run.create(ctx, env, s.Run); err != nil {
				return status, err
			}
		case runStateWait:
		}
	}

	// probe readiness
	if keepGoing, err := rt.readiness.reconcile(ctx, &env, s.Readiness, &status); err != nil || !keepGoing {
		return status, err
	}
	return status, nil
}
