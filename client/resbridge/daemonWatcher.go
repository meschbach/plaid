package resbridge

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/daemon"
	"github.com/meschbach/plaid/resources"
)

type daemonWatcher struct {
	w daemon.Watcher
	c chan resources.ResourceChanged
}

func (g *daemonWatcher) OnType(ctx context.Context, kind resources.Type, consume resources.OnResourceChanged) (resources.WatchToken, error) {
	return g.w.OnType(ctx, kind, consume)
}

func (g *daemonWatcher) OnResource(ctx context.Context, ref resources.Meta, consume resources.OnResourceChanged) (resources.WatchToken, error) {
	return g.w.OnResource(ctx, ref, consume)
}

func (g *daemonWatcher) Off(ctx context.Context, token resources.WatchToken) error {
	return g.w.Off(ctx, token)
}

func (g *daemonWatcher) Close(ctx context.Context) error {
	return g.w.Close(ctx)
}

func (g *daemonWatcher) Events() chan resources.ResourceChanged {
	return g.c
}

func (g *daemonWatcher) Digest(ctx context.Context, event resources.ResourceChanged) error {
	//TODO implement me
	panic("implement me")
}
