package resources

import "context"

type System interface {
	Storage(ctx context.Context) (Storage, error)
}

type Storage interface {
	Create(ctx context.Context, ref Meta, spec any, opts ...CreateOpt) error
	Delete(ctx context.Context, ref Meta) (bool, error)
	Get(ctx context.Context, ref Meta, spec any) (bool, error)
	GetStatus(ctx context.Context, ref Meta, status any) (bool, error)
	UpdateStatus(ctx context.Context, ref Meta, status any) (bool, error)
	GetEvents(ctx context.Context, ref Meta, level EventLevel) ([]Event, bool, error)
	Log(ctx context.Context, ref Meta, level EventLevel, fmt string, args ...any) (bool, error)
	List(ctx context.Context, kind Type) ([]Meta, error)
	Observer(ctx context.Context) (Watcher, error)
}

type Watcher interface {
	OnType(ctx context.Context, kind Type, consume OnResourceChanged) (WatchToken, error)
	OnResource(ctx context.Context, ref Meta, consume OnResourceChanged) (WatchToken, error)
	Off(ctx context.Context, token WatchToken) error
	Close(ctx context.Context) error
	Events() chan ResourceChanged
	Digest(ctx context.Context, event ResourceChanged) error
}
