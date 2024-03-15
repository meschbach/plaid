package dependencies

import (
	"context"
	"errors"
	"github.com/meschbach/plaid/resources"
	"go.opentelemetry.io/otel/codes"
)

// State manages a set of named dependencies
type State struct {
	deps map[string]*DependencyState
}

type Alpha1Spec []NamedDependencyAlpha1

type NamedDependencyAlpha1 struct {
	Name string         `json:"name"`
	Ref  resources.Meta `json:"ref"`
}

type Alpha1Status map[string]DependencyStatusAlpha1

type DependencyStatusAlpha1 struct {
	Name  string         `json:"name"`
	Ref   resources.Meta `json:"ref"`
	Ready bool           `json:"ready"`
}

func (s *State) Init(deps Alpha1Spec) {
	s.deps = make(map[string]*DependencyState)
	for _, d := range deps {
		s.deps[d.Name] = &DependencyState{
			ref: d.Ref,
		}
	}
}

// Reconcile updates the internal state and determines if all dependencies are ready
func (s *State) Reconcile(parent context.Context, env Env) (ready bool, status Alpha1Status, err error) {
	ctx, span := tracer.Start(parent, "dependencies.Reconcile")
	defer span.End()

	var allProblems []error
	allReady := true
	output := make(Alpha1Status)
	for name, dep := range s.deps {
		ready, problem := dep.Reconcile(ctx, env)
		output[name] = DependencyStatusAlpha1{
			Name:  name,
			Ref:   dep.ref,
			Ready: ready,
		}
		if problem != nil {
			span.SetStatus(codes.Error, "had problem reconciling dependency")
			allProblems = append(allProblems, problem)
			allReady = false
		} else {
			allReady = allReady && ready
		}
	}

	if !allReady {
		return false, output, errors.Join(allProblems...)
	}
	return allReady, output, nil
}
