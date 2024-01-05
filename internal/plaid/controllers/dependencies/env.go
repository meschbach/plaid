package dependencies

import (
	"context"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
)

type Env struct {
	Storage  *resources.Client
	Watcher  *resources.ClientWatcher
	OnChange func(ctx context.Context) error
}
