package client

import (
	"context"
	"github.com/meschbach/plaid/resources"
)

type daemonDispatch struct {
	event   resources.ResourceChanged
	consume resources.OnResourceChanged
}

// todo: this is a tangled mess which needs to be cleaned up
type daemonWatcher struct {
	underlyingWatcher Watcher
	feed              chan resources.ResourceChanged
	//todo: feed really needs an opaque type
	realFeed chan daemonDispatch
}

func (g *daemonWatcher) OnType(ctx context.Context, kind resources.Type, consume resources.OnResourceChanged) (resources.WatchToken, error) {
	return g.underlyingWatcher.OnType(ctx, kind, g.pushOnQueue(consume))
}

func (g *daemonWatcher) OnResource(ctx context.Context, ref resources.Meta, consume resources.OnResourceChanged) (resources.WatchToken, error) {
	return g.underlyingWatcher.OnResource(ctx, ref, g.pushOnQueue(consume))
}

func (g *daemonWatcher) OnResourceStatusChanged(ctx context.Context, ref resources.Meta, consume resources.OnResourceChanged) (resources.WatchToken, error) {
	invoke := g.pushOnQueue(consume)
	return g.underlyingWatcher.OnResource(ctx, ref, func(ctx context.Context, changed resources.ResourceChanged) error {
		switch changed.Operation {
		case resources.StatusUpdated:
			return invoke(ctx, changed)
		case resources.DeletedEvent:
			return invoke(ctx, changed)
		default:
			return nil
		}
	})
}

func (g *daemonWatcher) Off(ctx context.Context, token resources.WatchToken) error {
	return g.underlyingWatcher.Off(ctx, token)
}

func (g *daemonWatcher) Close(ctx context.Context) error {
	return g.underlyingWatcher.Close(ctx)
}

func (g *daemonWatcher) Events() chan resources.ResourceChanged {
	return g.feed
}

func (g *daemonWatcher) Digest(ctx context.Context, event resources.ResourceChanged) error {
	//slippery slope, I know!
	select {
	case dispatch := <-g.realFeed:
		err := dispatch.consume(ctx, dispatch.event)
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (g *daemonWatcher) pushOnQueue(consume resources.OnResourceChanged) resources.OnResourceChanged {
	return func(ctx context.Context, changed resources.ResourceChanged) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case g.realFeed <- daemonDispatch{
			consume: consume,
			event:   changed,
		}:
			select {
			case g.feed <- changed:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}
