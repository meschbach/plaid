package service

import (
	"context"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources/operator"
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
