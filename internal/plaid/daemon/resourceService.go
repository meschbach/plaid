package daemon

import (
	"context"
	"fmt"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/daemon/wire"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"github.com/thejerf/suture/v4"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ResourceService struct {
	wire.UnimplementedResourceControllerServer
	client *resources.Client
}

func (d *ResourceService) Create(ctx context.Context, w *wire.CreateResourceIn) (*wire.CreateResourceOut, error) {
	err := d.client.CreateBytes(ctx, internalizeMeta(w.Target), w.Spec)
	return &wire.CreateResourceOut{}, err
}

func (d *ResourceService) Get(ctx context.Context, in *wire.GetIn) (*wire.GetOut, error) {
	data, exists, err := d.client.GetBytes(ctx, internalizeMeta(in.Target))
	return &wire.GetOut{
		Exists: exists,
		Spec:   data,
	}, err
}

func (d *ResourceService) GetStatus(ctx context.Context, in *wire.GetStatusIn) (*wire.GetStatusOut, error) {
	data, exists, err := d.client.GetStatusBytes(ctx, internalizeMeta(in.Target))
	return &wire.GetStatusOut{
		Exists: exists,
		Status: data,
	}, err
}

func (d *ResourceService) GetEvents(ctx context.Context, in *wire.GetEventsIn) (*wire.GetEventsOut, error) {
	events, exists, err := d.client.GetLogs(ctx, internalizeMeta(in.Ref), internalizeEventLevel(in.Level))
	if err != nil {
		return nil, err
	}

	out := &wire.GetEventsOut{
		Exists: exists,
	}
	out.Events = make([]*wire.Event, len(events))
	for i, e := range events {
		out.Events[i] = &wire.Event{
			When:     timestamppb.New(e.When),
			Level:    externalizeEventLevel(e.Level),
			Rendered: fmt.Sprintf(e.Format, e.Params...),
		}
	}
	return out, nil
}

func (d *ResourceService) List(ctx context.Context, in *wire.ListIn) (*wire.ListOut, error) {
	kind := internalizeType(in.Type)
	refs, err := d.client.List(ctx, kind)
	if err != nil {
		return nil, err
	}
	result := &wire.ListOut{
		Ref: make([]*wire.Meta, len(refs)),
	}
	for i, r := range refs {
		result.Ref[i] = metaToWire(r)
	}
	return result, nil
}

func internalizeType(p *wire.Type) resources.Type {
	return resources.Type{
		Kind:    p.Kind,
		Version: p.Version,
	}
}
func internalizeMeta(meta *wire.Meta) resources.Meta {
	return resources.Meta{
		Type: internalizeType(meta.Kind),
		Name: meta.Name,
	}
}

func (d *ResourceService) Watcher(w wire.ResourceController_WatcherServer) error {
	ctx, span := tracer.Start(w.Context(), "ResourceService.Watcher: implementation")
	defer span.End()
	watcher, err := d.client.Watcher(ctx)
	if err != nil {
		return err
	}

	events := make(chan *wire.WatcherEventIn, 4)
	pump := &serviceWatcherInputPump{
		stream: w,
		events: events,
	}
	bridge := &watcherBridge{
		stream:  w,
		events:  events,
		watcher: watcher,
		tokens:  make(map[uint64]resources.WatchToken),
	}
	s := suture.NewSimple("watcher")
	s.Add(pump)
	s.Add(bridge)
	result := s.Serve(w.Context())
	if result != nil {
		if result == context.Canceled {
			result = nil
		} else {
			span.SetStatus(codes.Error, "watcher service failed")
			span.RecordError(err)
		}
	}
	return result
}
