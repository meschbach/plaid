package resources

import "context"

type System interface {
	Storage(ctx context.Context) (Storage, error)
}

type Storage interface {
	Create(ctx context.Context, ref Meta, spec any, opts ...CreateOpt) error
	Delete(ctx context.Context, ref Meta) (bool, error)
	Get(ctx context.Context, ref Meta, spec any) (bool, error)
	//Update updates the specification for the given ref and propagates events out.
	Update(ctx context.Context, ref Meta, spec any) (exists bool, problem error)

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

	//OnResourceStatusChanged registers an event handler when either the status is changed or the resource is deleted.
	//This allows for optimizations within the watcher subsystem to only observe a subset of event.
	//todo: reconsider if this is better in a wrapper or at a different level. quickly explodes in complexity
	OnResourceStatusChanged(ctx context.Context, ref Meta, consume OnResourceChanged) (WatchToken, error)
	Off(ctx context.Context, token WatchToken) error
	Close(ctx context.Context) error
	Events() chan ResourceChanged
	Digest(ctx context.Context, event ResourceChanged) error
}
