package dependencies

import (
	"context"
	"errors"
	"github.com/meschbach/plaid/internal/plaid/resources"
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
	Name  string `json:"name"`
	Ready bool   `json:"ready"`
}

func (s *State) Init(deps Alpha1Spec) {
	s.deps = make(map[string]*DependencyState)
	for _, d := range deps {
		s.deps[d.Name] = &DependencyState{
			ref: d.Ref,
		}
	}
}

func (s *State) Reconcile(ctx context.Context, env Env) (ready bool, status Alpha1Status, err error) {
	var allProblems []error
	allReady := true
	output := make(Alpha1Status)
	for name, dep := range s.deps {
		ready, problem := dep.Reconcile(ctx, env)
		output[name] = DependencyStatusAlpha1{
			Name:  name,
			Ready: ready,
		}
		if problem != nil {
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
