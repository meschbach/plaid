package service

import (
	"context"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/operator"
)

type serviceState struct {
	bridge *operator.KindBridgeState

	dependencies []*dependencyState
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
