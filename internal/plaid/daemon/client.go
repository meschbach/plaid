package daemon

import (
	"context"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
)

type Client interface {
	Create(ctx context.Context, ref resources.Meta, spec any) error
	Get(ctx context.Context, ref resources.Meta, spec any) (bool, error)
	GetStatus(ctx context.Context, ref resources.Meta, status any) (bool, error)
	GetEvents(ctx context.Context, ref resources.Meta, level resources.EventLevel) ([]resources.Event, error)
	List(ctx context.Context, kind resources.Type) ([]resources.Meta, error)
	Watcher(ctx context.Context) (Watcher, error)
}

type Watcher interface {
	OnType(ctx context.Context, kind resources.Type, consume resources.OnResourceChanged) (resources.WatchToken, error)
	OnResource(ctx context.Context, ref resources.Meta, consume resources.OnResourceChanged) (resources.WatchToken, error)
	Off(ctx context.Context, token resources.WatchToken) error
	Close(ctx context.Context) error
}
