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
	//dumb but works until I figured out a better API
	realFeed chan daemonDispatch
}

func (g *daemonWatcher) OnType(ctx context.Context, kind resources.Type, consume resources.OnResourceChanged) (resources.WatchToken, error) {
	return g.underlyingWatcher.OnType(ctx, kind, g.pushOnQueue(consume))
}

func (g *daemonWatcher) OnResource(ctx context.Context, ref resources.Meta, consume resources.OnResourceChanged) (resources.WatchToken, error) {
	return g.underlyingWatcher.OnResource(ctx, ref, g.pushOnQueue(consume))
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
	dispatch := <-g.realFeed
	return dispatch.consume(ctx, dispatch.event)
}

func (g *daemonWatcher) pushOnQueue(consume resources.OnResourceChanged) resources.OnResourceChanged {
	return func(ctx context.Context, changed resources.ResourceChanged) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case g.feed <- changed:
			g.realFeed <- daemonDispatch{
				consume: consume,
				event:   changed,
			}
			return nil
		}
	}
}
