package tooling

import (
	"context"
	"github.com/meschbach/plaid/resources"
)

type Env struct {
	// Subject is the object the controller is currently operating on.
	Subject   resources.Meta
	Storage   resources.Storage
	Watcher   resources.Watcher
	Reconcile func(ctx context.Context) error
}
