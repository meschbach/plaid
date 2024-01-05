package daemon

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/internal/plaid/daemon/wire"
	"github.com/meschbach/plaid/resources"
)

// watcherBridge is the service side drain for a specific watch.
type watcherBridge struct {
	events  <-chan *wire.WatcherEventIn
	stream  wire.ResourceController_WatcherServer
	watcher *resources.ClientWatcher
	tokens  map[uint64]resources.WatchToken
}

func (w *watcherBridge) Serve(ctx context.Context) error {
	for {
		select {
		case event := <-w.events:
			if err := w.consumeInput(ctx, event); err != nil {
				return err
			}
		case event := <-w.watcher.Feed:
			if err := w.watcher.Digest(ctx, event); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (w *watcherBridge) consumeInput(ctx context.Context, e *wire.WatcherEventIn) error {
	//Client requested a resource watch
	if e.OnResource != nil {
		res := internalizeMeta(e.OnResource)
		token, err := w.watcher.OnResource(ctx, res, func(ctx context.Context, changed resources.ResourceChanged) error {
			return w.stream.Send(&wire.WatcherEventOut{
				Tag: e.Tag,
				Ref: e.OnResource,
				Op:  exportOp(changed.Operation),
			})
		})
		if err != nil {
			return err
		}
		w.tokens[e.Tag] = token
	}
	if e.OnType != nil {
		token, err := w.watcher.OnType(ctx, internalizeType(e.OnType), func(ctx context.Context, changed resources.ResourceChanged) error {
			return w.stream.Send(&wire.WatcherEventOut{
				Tag: e.Tag,
				Ref: e.OnResource,
				Op:  exportOp(changed.Operation),
			})
		})
		if err != nil {
			return err
		}
		w.tokens[e.Tag] = token
	}
	if e.Delete != nil && *e.Delete {
		if t, has := w.tokens[e.Tag]; has {
			if err := w.watcher.Off(ctx, t); err != nil {
				return err
			}
			delete(w.tokens, e.Tag)
		} else {
			fmt.Printf("[grpc.watcherBridge] WARNING: wire tag %d does not have an internal analog to delete\n", e.Tag)
		}
	}
	return nil
}

func exportOp(operation resources.ResourceChangedOperation) wire.WatcherEventOut_Op {
	switch operation {
	case resources.CreatedEvent:
		return wire.WatcherEventOut_Created
	case resources.StatusUpdated:
		return wire.WatcherEventOut_UpdatedStatus
	default:
		panic(fmt.Sprintf("unknown operation %d", operation))
	}
}
