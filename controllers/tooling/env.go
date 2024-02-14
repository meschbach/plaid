package tooling

import (
	"context"
	"github.com/meschbach/plaid/resources"
)

type Env struct {
	// Subject is the object the controller is currently operating on.
	Subject   resources.Meta
	Storage   *resources.Client
	Watcher   *resources.ClientWatcher
	Reconcile func(ctx context.Context) error
}
