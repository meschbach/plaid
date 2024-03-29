package service

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/ipc/grpc/reswire"
	"github.com/meschbach/plaid/resources"
)

// watcherBridge is the service side drain for a specific watch.
type watcherBridge struct {
	events  <-chan *reswire.WatcherEventIn
	stream  reswire.ResourceController_WatcherServer
	watcher *resources.ClientWatcher
	tokens  map[uint64]resources.WatchToken
}

func (w *watcherBridge) Serve(ctx context.Context) error {
	for {
		select {
		case event, ok := <-w.events:
			//underlyingWatcher.events was closed
			if !ok {
				continue
			}
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

func (w *watcherBridge) consumeInput(ctx context.Context, e *reswire.WatcherEventIn) error {
	//Client requested a resource watch
	if e.OnResource != nil {
		res := internalizeMeta(e.OnResource)
		token, err := w.watcher.OnResource(ctx, res, func(ctx context.Context, changed resources.ResourceChanged) error {
			return w.stream.Send(&reswire.WatcherEventOut{
				Tag: e.Tag,
				Ref: e.OnResource,
				Op:  exportOp(changed.Operation),
			})
		})
		if err != nil {
			return err
		}
		w.tokens[e.Tag] = token
		if err := w.stream.Send(&reswire.WatcherEventOut{
			Tag: e.Tag,
			Op:  reswire.WatcherEventOut_ChangeAck,
		}); err != nil {
			return err
		}
	}
	if e.OnType != nil {
		token, err := w.watcher.OnType(ctx, reswire.InternalizeKind(e.OnType), func(ctx context.Context, changed resources.ResourceChanged) error {
			ref := reswire.MetaToWire(changed.Which)
			return w.stream.Send(&reswire.WatcherEventOut{
				Tag: e.Tag,
				Ref: ref,
				Op:  exportOp(changed.Operation),
			})
		})
		if err != nil {
			return err
		}
		//todo: sync table
		w.tokens[e.Tag] = token
		//send acknowledgement of registration
		if err := w.stream.Send(&reswire.WatcherEventOut{
			Tag: e.Tag,
			Op:  reswire.WatcherEventOut_ChangeAck,
		}); err != nil {
			return err
		}
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

func exportOp(operation resources.ResourceChangedOperation) reswire.WatcherEventOut_Op {
	switch operation {
	case resources.CreatedEvent:
		return reswire.WatcherEventOut_Created
	case resources.StatusUpdated:
		return reswire.WatcherEventOut_UpdatedStatus
	case resources.DeletedEvent:
		return reswire.WatcherEventOut_Deleted
	default:
		panic(fmt.Sprintf("unknown operation %d", operation))
	}
}
