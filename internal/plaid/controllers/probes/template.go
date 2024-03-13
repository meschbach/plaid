package probes

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/httpProbe"
	"github.com/meschbach/plaid/resources"
)

type TemplateEnv struct {
	ClaimedBy resources.Meta
	Storage   resources.Storage
	Watcher   resources.Watcher
	OnChange  func(ctx context.Context) error
}

type TemplateAlpha1Spec struct {
	Http *httpProbe.TemplateAlpha1 `json:"http,omitempty"`
}

// todo: revisit and move env to tooling.Env
func (t *TemplateAlpha1Spec) Instantiate(ctx context.Context, env TemplateEnv) (TemplateAlpha1State, error) {
	if t == nil {
		return &alwaysReady{}, nil
	}
	if t.Http != nil {
		state, err := t.Http.Instantiate(ctx, env.Storage, env.ClaimedBy, env.Watcher, env.OnChange)
		return state, err
	}
	return &alwaysReady{}, nil
}

type TemplateAlpha1State interface {
	Reconcile(ctx context.Context, storage resources.Storage) error
	Ready() bool
}

type alwaysReady struct{}

func (a *alwaysReady) Reconcile(ctx context.Context, storage resources.Storage) error {
	return nil
}

func (a *alwaysReady) Ready() bool {
	return true
}
