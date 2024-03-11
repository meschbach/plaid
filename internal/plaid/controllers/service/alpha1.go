package service

import (
	"context"
	"errors"
	"github.com/meschbach/plaid/controllers/tooling"
	"github.com/meschbach/plaid/internal/plaid/controllers/dependencies"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/internal/plaid/controllers/probes"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/operator"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"sort"
	"strings"
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
	// When RestartToken changes a new build will occur (if one exists) and if successful will stop the existing process
	// and start the new one.
	RestartToken string `json:"restart-token"`
}

type Alpha1Status struct {
	Dependencies []Alpha1StatusDependency `json:"dependencies,omitempty"`
	Build        Alpha1BuildStatus        `json:"build,omitempty"`
	Run          Alpha1RunStatus          `json:"run"`
	Ready        bool                     `json:"ready"`
	RunningToken string                   `json:"running-token"`
}

type Alpha1StatusDependency struct {
	Dependency resources.Meta `json:"ref"`
	Ready      bool           `json:"ready"`
}

type Alpha1BuildStatus struct {
	State string          `json:"state"`
	Ref   *resources.Meta `json:"ref,omitempty"`
}

const StateNotReady = "not-ready"
const Running = "running"

type Alpha1RunStatus struct {
	State string          `json:"state"`
	Ref   *resources.Meta `json:"ref,omitempty"`
}

type alpha1Ops struct {
	client  *resources.Client
	watcher *resources.ClientWatcher
}

func (a *alpha1Ops) Create(ctx context.Context, which resources.Meta, spec Alpha1Spec, bridge *operator.KindBridgeState) (*serviceState, Alpha1Status, error) {
	token := resources.GenSuffix(4)
	rt := &serviceState{
		token:        token,
		bridge:       bridge,
		dependencies: &dependencies.State{},
	}
	deps := make([]dependencies.NamedDependencyAlpha1, 0, len(spec.Dependencies))
	for _, ref := range spec.Dependencies {
		deps = append(deps, dependencies.NamedDependencyAlpha1{
			Name: ref.Name,
			Ref:  ref,
		})
	}
	rt.dependencies.Init(deps)

	status, err := a.Update(ctx, which, rt, spec)
	return rt, status, err
}

func (a *alpha1Ops) Update(parent context.Context, which resources.Meta, rt *serviceState, s Alpha1Spec) (Alpha1Status, error) {
	ctx, span := tracer.Start(parent, "service.alpha1/Update", trace.WithAttributes(which.AsTraceAttribute("which")...))
	defer span.End()

	env := tooling.Env{
		Subject: which,
		Storage: a.client,
		Watcher: a.watcher,
		Reconcile: func(ctx context.Context) error {
			return rt.bridge.OnResourceChange(ctx, which)
		},
	}

	status := Alpha1Status{}
	status.Run = rt.run.toStatus()
	//todo: alpha2 should just use the status directly
	allReady, depStatus, err := rt.dependencies.Reconcile(ctx, dependencies.Env{
		Storage: a.client,
		Watcher: a.watcher,
		OnChange: func(ctx context.Context) error {
			return env.Reconcile(ctx)
		},
	})
	for _, s := range depStatus {
		status.Dependencies = append(status.Dependencies, Alpha1StatusDependency{
			Dependency: s.Ref,
			Ready:      s.Ready,
		})
	}
	//todo: figure out how to test stable output
	sort.SliceStable(status.Dependencies, func(i, j int) bool {
		lhs := status.Dependencies[i]
		rhs := status.Dependencies[j]
		return strings.Compare(lhs.Dependency.Name, rhs.Dependency.Name) > 0
	})

	if err != nil {
		span.SetStatus(codes.Error, "dependency reconciliation error")
		span.RecordError(err)
		return status, err
	}
	if !allReady {
		span.AddEvent("dependencies-not-ready")
		return status, nil
	}

	//setup build
	if s.Build != nil { //todo: test builder branch
		if step, buildStatus, err := rt.build.decideNextStep(ctx, env); err != nil {
			span.SetStatus(codes.Error, "failed to decide next steps")
			return status, err
		} else {
			status.Build = buildStatus
			switch step {
			case builderNextCreate:
				if err := rt.build.create(ctx, env, s.Build, &status.Build); err != nil {
					span.SetStatus(codes.Error, "build error")
					status.Build.State = "internal-error"
				}
				return status, err
			case builderNextWait:
				span.AddEvent("builder-wait")
				return status, err
			case builderStateSuccessfullyCompleted:
				//continue
			}
		}
	}

	//run service
	if step, err := rt.run.decideNextStep(ctx, env); err != nil {
		span.SetStatus(codes.Error, "runtime error")
		return status, err
	} else {
		status.Run = rt.run.toStatus()
		switch step {
		case runStateCreate:
			span.AddEvent("run-create")
			if err := rt.run.create(ctx, env, s.Run); err != nil {
				span.SetStatus(codes.Error, "runtime create error")
				return status, err
			}
			status.Run = rt.run.toStatus()
			return status, nil
		case runStateWait:
			span.AddEvent("run-wait")
			return status, err
		case runStateRunning:
			span.AddEvent("run-running")
		}
	}

	// probe readiness
	_, err = rt.readiness.reconcile(ctx, env, s.Readiness, &status)
	if err != nil {
		status.Ready = false
		span.SetStatus(codes.Error, "readiness reconciliation")
		return status, err
	}
	return status, nil
}

func (a *alpha1Ops) Delete(ctx context.Context, which resources.Meta, rt *serviceState) error {
	env := tooling.Env{
		Subject: which,
		Storage: a.client,
		Watcher: a.watcher,
		Reconcile: func(ctx context.Context) error {
			return rt.bridge.OnResourceChange(ctx, which)
		},
	}

	buildError := rt.build.delete(ctx, env)
	runError := rt.run.delete(ctx, env)
	return errors.Join(buildError, runError)
}
