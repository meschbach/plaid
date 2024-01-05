package projectfile

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/resources"
	"github.com/meschbach/plaid/internal/plaid/resources/operator"
	"time"
)

var Alpha1 = resources.Type{
	Kind:    Kind,
	Version: "alpha1",
}

type Alpha1Spec struct {
	WorkingDirectory string `json:"wd"`
	//ProjectFile must be relative to the working directory for the project
	ProjectFile string `json:"project-file"`
}

type Alpha1Status struct {
	LastLoaded   *time.Time      `json:"last-loaded"`
	ProjectFile  string          `json:"project-file"`
	ParsingError *string         `json:"parsing-error,omitempty"`
	Current      *resources.Meta `json:"current-project,omitempty"`

	//Done is a legacy field originating from local-project.plaid.meschbach.com used to signal to clients all positive
	//terminal conditions have been met.  Typically this means all one shots have completed successfully
	Done    bool `json:"done"`
	Success bool `json:"success"`
}

type alpha1Ops struct {
	storage *resources.Client
	watcher *resources.ClientWatcher
}

func (a *alpha1Ops) Create(ctx context.Context, which resources.Meta, spec Alpha1Spec, bridge *operator.KindBridgeState) (*state, Alpha1Status, error) {
	s := &state{
		refresh: bridge,
	}

	status, err := a.Update(ctx, which, s, spec)
	return s, status, err
}

func (a *alpha1Ops) Update(ctx context.Context, which resources.Meta, rt *state, s Alpha1Spec) (Alpha1Status, error) {
	status := Alpha1Status{}
	rt.projectFile.updateStatus(s, &status)
	rt.projectResource.updateStatus(ctx, &status)
	if step, err := rt.projectFile.decideNextStep(ctx); err != nil {
		return status, err
	} else {
		switch step {
		case fileStateParse:
			if err := rt.projectFile.parse(ctx, s); err != nil {
				rt.projectFile.updateStatus(s, &status)
				return status, nil
			}
		case fileStateDone: //do nothing, expected typical state
		}
	}

	env := stateEnv{
		which:   which,
		rpc:     a.storage,
		watcher: a.watcher,
		reconcile: func(ctx context.Context) error {
			return rt.refresh.OnResourceChange(ctx, which)
		},
	}

	if step, err := rt.projectResource.decideNextSteps(ctx, env); err != nil {
		return status, err
	} else {
		switch step {
		case projectNextWait:
			return status, nil
		case projectNextCreate:
			err := rt.projectResource.create(ctx, env, s, rt.projectFile.parsedFile)
			rt.projectResource.updateStatus(ctx, &status)
			return status, err
		case projectNextDone:
		}
		rt.projectResource.updateStatus(ctx, &status)
	}
	return status, nil
}

type state struct {
	refresh         *operator.KindBridgeState
	projectFile     fileState
	projectResource projectState
}
