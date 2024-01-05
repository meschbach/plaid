package daemon

import (
	"context"
	"encoding/json"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/daemon/wire"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"github.com/thejerf/suture/v4"
)

type WireClientAdapter struct {
	wire     wire.ResourceControllerClient
	workPool *suture.Supervisor
}

func NewWireClientAdapter(workPool *suture.Supervisor, client wire.ResourceControllerClient) *WireClientAdapter {
	return &WireClientAdapter{
		wire:     client,
		workPool: workPool,
	}
}

func (w *WireClientAdapter) Create(ctx context.Context, ref resources.Meta, spec any) error {
	bytes, err := json.Marshal(spec)
	if err != nil {
		return err
	}

	wireRef := metaToWire(ref)
	_, err = w.wire.Create(ctx, &wire.CreateResourceIn{
		Target: wireRef,
		Spec:   bytes,
	})
	return err
}

func (w *WireClientAdapter) Get(ctx context.Context, ref resources.Meta, spec any) (bool, error) {
	wireRef := metaToWire(ref)
	out, err := w.wire.Get(ctx, &wire.GetIn{
		Target: wireRef,
	})
	if err != nil {
		return false, err
	}
	if !out.Exists {
		return false, nil
	}

	return true, json.Unmarshal(out.Spec, spec)
}

func (w *WireClientAdapter) GetStatus(ctx context.Context, ref resources.Meta, status any) (bool, error) {
	wireRef := metaToWire(ref)
	out, err := w.wire.GetStatus(ctx, &wire.GetStatusIn{
		Target: wireRef,
	})
	if err != nil {
		return false, err
	}
	if !out.Exists {
		return false, nil
	}
	if len(out.Status) == 0 {
		return true, nil
	}

	return true, json.Unmarshal(out.Status, status)
}

func (w *WireClientAdapter) GetEvents(ctx context.Context, ref resources.Meta, level resources.EventLevel) ([]resources.Event, error) {
	wireRef := metaToWire(ref)
	out, err := w.wire.GetEvents(ctx, &wire.GetEventsIn{
		Ref:   wireRef,
		Level: externalizeEventLevel(level),
	})
	if err != nil {
		return nil, err
	}
	if !out.Exists {
		return nil, nil
	}

	internalized := make([]resources.Event, len(out.Events))
	for i, e := range out.Events {
		internalized[i] = resources.Event{
			When:   e.When.AsTime(),
			Level:  internalizeEventLevel(e.Level),
			Format: e.Rendered,
			Params: nil,
		}
	}
	return internalized, nil
}

func (w *WireClientAdapter) List(ctx context.Context, kind resources.Type) ([]resources.Meta, error) {
	wireKind := typeToWire(kind)
	out, err := w.wire.List(ctx, &wire.ListIn{Type: wireKind})
	if err != nil {
		return nil, err
	}
	result := make([]resources.Meta, len(out.Ref))
	for i, r := range out.Ref {
		result[i] = internalizeMeta(r)
	}
	return result, err
}

func (w *WireClientAdapter) Watcher(ctx context.Context) (Watcher, error) {
	wc, err := w.wire.Watcher(ctx)
	if err != nil {
		return nil, err
	}
	adapter := &watcherAdapter{
		wire:   w.wire,
		stream: wc,
		tags:   make(map[resources.WatchToken]*wireAdapterHandler),
	}
	w.workPool.Add(adapter)
	return adapter, nil
}
