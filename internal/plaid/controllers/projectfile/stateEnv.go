package projectfile

import (
	"context"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
)

type stateEnv struct {
	which     resources.Meta
	rpc       *resources.Client
	watcher   *resources.ClientWatcher
	reconcile func(ctx context.Context) error
}
