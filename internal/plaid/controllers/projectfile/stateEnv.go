package projectfile

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/resources"
)

type stateEnv struct {
	which     resources.Meta
	rpc       *resources.Client
	watcher   *resources.ClientWatcher
	reconcile func(ctx context.Context) error
}
