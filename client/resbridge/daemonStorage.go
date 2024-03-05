package resbridge

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/daemon"
	"github.com/meschbach/plaid/internal/plaid/daemon/wire"
	"github.com/meschbach/plaid/resources"
)

type daemonStorage struct {
	client daemon.Client
	wire   wire.ResourceControllerClient
}

// todo: implement creation options
func (g *daemonStorage) Create(ctx context.Context, ref resources.Meta, spec any, opts ...resources.CreateOpt) error {
	if len(opts) > 0 {
		panic("todo")
	}
	return g.client.Create(ctx, ref, spec)
}

// todo: chain through exists result
func (g *daemonStorage) Delete(ctx context.Context, ref resources.Meta) (bool, error) {
	e := g.client.Delete(ctx, ref)
	return true, e
}

func (g *daemonStorage) Get(ctx context.Context, ref resources.Meta, spec any) (bool, error) {
	return g.client.Get(ctx, ref, spec)
}

func (g daemonStorage) GetStatus(ctx context.Context, ref resources.Meta, status any) (bool, error) {
	return g.client.GetStatus(ctx, ref, status)
}

func (g daemonStorage) UpdateStatus(ctx context.Context, ref resources.Meta, status any) (bool, error) {
	panic("todo")
}

func (g daemonStorage) GetEvents(ctx context.Context, ref resources.Meta, level resources.EventLevel) ([]resources.Event, bool, error) {
	//TODO implement me
	panic("implement me")
}

func (g daemonStorage) Log(ctx context.Context, ref resources.Meta, level resources.EventLevel, fmt string, args ...any) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (g *daemonStorage) List(ctx context.Context, kind resources.Type) ([]resources.Meta, error) {
	return g.client.List(ctx, kind)
}

func (g *daemonStorage) Observer(ctx context.Context) (resources.Watcher, error) {
	w, err := g.client.Watcher(ctx)
	if err != nil {
		return nil, err
	}
	c := make(chan resources.ResourceChanged)
	return &daemonWatcher{w, c}, nil
}
