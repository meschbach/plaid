package service

import (
	"context"
	"github.com/meschbach/plaid/controllers/tooling"
	"github.com/meschbach/plaid/internal/plaid/controllers/dependencies"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/operator"
)

type serviceState struct {
	bridge *operator.KindBridgeState

	dependencies *dependencies.State
	build        builderState
	run          runState
	readiness    readinessProbeState
}

type resEnv struct {
	object    resources.Meta
	rpc       *resources.Client
	watcher   *resources.ClientWatcher
	reconcile func(ctx context.Context) error
}

func (r resEnv) toTooling() tooling.Env {
	return tooling.Env{
		Subject:   r.object,
		Storage:   r.rpc,
		Watcher:   r.watcher,
		Reconcile: r.reconcile,
	}
}
