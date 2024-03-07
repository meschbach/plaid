package client

import (
	"context"
	"encoding/json"
	"github.com/meschbach/plaid/ipc/grpc/reswire"
	"github.com/meschbach/plaid/resources"
	"github.com/thejerf/suture/v4"
	"sync"
)

type WireClientAdapter struct {
	wire     reswire.ResourceControllerClient
	workPool *suture.Supervisor
}

func NewWireClientAdapter(workPool *suture.Supervisor, client reswire.ResourceControllerClient) *WireClientAdapter {
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

	wireRef := reswire.MetaToWire(ref)
	_, err = w.wire.Create(ctx, &reswire.CreateResourceIn{
		Target: wireRef,
		Spec:   bytes,
	})
	return err
}

func (w *WireClientAdapter) Delete(ctx context.Context, ref resources.Meta) error {
	wireRef := reswire.MetaToWire(ref)
	_, err := w.wire.Delete(ctx, &reswire.DeleteResourceIn{Ref: wireRef})
	return err
}

func (w *WireClientAdapter) Get(ctx context.Context, ref resources.Meta, spec any) (bool, error) {
	wireRef := reswire.MetaToWire(ref)
	out, err := w.wire.Get(ctx, &reswire.GetIn{
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
	wireRef := reswire.MetaToWire(ref)
	out, err := w.wire.GetStatus(ctx, &reswire.GetStatusIn{
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
	wireRef := reswire.MetaToWire(ref)
	out, err := w.wire.GetEvents(ctx, &reswire.GetEventsIn{
		Ref:   wireRef,
		Level: reswire.ExternalizeEventLevel(level),
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
			Level:  reswire.InternalizeEventLevel(e.Level),
			Format: e.Rendered,
			Params: nil,
		}
	}
	return internalized, nil
}

func (w *WireClientAdapter) List(ctx context.Context, kind resources.Type) ([]resources.Meta, error) {
	wireKind := reswire.ExternalizeType(kind)
	out, err := w.wire.List(ctx, &reswire.ListIn{Type: wireKind})
	if err != nil {
		return nil, err
	}
	result := make([]resources.Meta, len(out.Ref))
	for i, r := range out.Ref {
		result[i] = reswire.InternalizeMeta(r)
	}
	return result, err
}

func (w *WireClientAdapter) Watcher(ctx context.Context) (Watcher, error) {
	wc, err := w.wire.Watcher(ctx)
	if err != nil {
		return nil, err
	}
	adapter := &watcherAdapter{
		wire:     w.wire,
		stream:   wc,
		tags:     make(map[resources.WatchToken]*wireAdapterHandler),
		ackTable: make(map[uint64]*reswire.WatcherEventOut),
		ackLock:  &sync.Mutex{},
	}
	adapter.ackCondition = sync.NewCond(adapter.ackLock)
	w.workPool.Add(adapter)
	return adapter, nil
}
