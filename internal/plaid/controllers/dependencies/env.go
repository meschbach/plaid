package dependencies

import (
	"context"
	"github.com/meschbach/plaid/resources"
)

type Env struct {
	Storage  *resources.Client
	Watcher  *resources.ClientWatcher
	OnChange func(ctx context.Context) error
}
