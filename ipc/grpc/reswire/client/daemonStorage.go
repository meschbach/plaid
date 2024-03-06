package client

import (
	"context"
	"encoding/json"
	"github.com/meschbach/plaid/ipc/grpc/reswire"
	"github.com/meschbach/plaid/resources"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"time"
)

type daemonStorage struct {
	client Client
	wire   reswire.ResourceControllerClient
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

func (g *daemonStorage) GetStatus(ctx context.Context, ref resources.Meta, status any) (bool, error) {
	return g.client.GetStatus(ctx, ref, status)
}

func (g *daemonStorage) UpdateStatus(ctx context.Context, ref resources.Meta, status any) (bool, error) {
	asBytes, err := json.Marshal(status)
	if err != nil {
		return false, err
	}

	out, err := g.wire.UpdateStatus(ctx, &reswire.UpdateStatusIn{
		Target: reswire.MetaToWire(ref),
		Status: asBytes,
	})
	if err != nil {
		return false, err
	}
	return out.Exists, err
}

func (g *daemonStorage) GetEvents(ctx context.Context, ref resources.Meta, level resources.EventLevel) ([]resources.Event, bool, error) {
	events, err := g.client.GetEvents(ctx, ref, level)
	return events, true, err
}

func (g *daemonStorage) Log(ctx context.Context, ref resources.Meta, level resources.EventLevel, fmt string, args ...any) (bool, error) {
	out, err := g.wire.Log(ctx, &reswire.LogIn{
		Ref:   reswire.MetaToWire(ref),
		Event: reswire.Eventf(time.Now(), level, fmt, args...),
	})
	if err != nil {
		span := trace.SpanFromContext(ctx)
		span.SetStatus(codes.Error, "failed to log")
		span.RecordError(err)
		return false, err
	}
	return out.Exists, err
}

func (g *daemonStorage) List(ctx context.Context, kind resources.Type) ([]resources.Meta, error) {
	return g.client.List(ctx, kind)
}

func (g *daemonStorage) Observer(ctx context.Context) (resources.Watcher, error) {
	w, err := g.client.Watcher(ctx)
	if err != nil {
		return nil, err
	}
	c := make(chan resources.ResourceChanged, 10)
	realFeed := make(chan daemonDispatch, 10)
	return &daemonWatcher{
		underlyingWatcher: w,
		feed:              c,
		realFeed:          realFeed,
	}, nil
}
