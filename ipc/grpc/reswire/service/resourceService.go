package service

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/ipc/grpc/reswire"
	"github.com/meschbach/plaid/resources"
	"github.com/thejerf/suture/v4"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ResourceService struct {
	reswire.UnimplementedResourceControllerServer
	client *resources.Client
}

func (d *ResourceService) Create(ctx context.Context, w *reswire.CreateResourceIn) (*reswire.CreateResourceOut, error) {
	err := d.client.CreateBytes(ctx, reswire.InternalizeMeta(w.Target), w.Spec)
	return &reswire.CreateResourceOut{}, err
}

func (d *ResourceService) Delete(ctx context.Context, w *reswire.DeleteResourceIn) (*reswire.DeleteResourceOut, error) {
	target := reswire.InternalizeMeta(w.Ref)
	exists, err := d.client.Delete(ctx, target)
	return &reswire.DeleteResourceOut{Success: exists}, err
}

func (d *ResourceService) Get(ctx context.Context, in *reswire.GetIn) (*reswire.GetOut, error) {
	data, exists, err := d.client.GetBytes(ctx, reswire.InternalizeMeta(in.Target))
	return &reswire.GetOut{
		Exists: exists,
		Spec:   data,
	}, err
}

func (d *ResourceService) GetStatus(ctx context.Context, in *reswire.GetStatusIn) (*reswire.GetStatusOut, error) {
	data, exists, err := d.client.GetStatusBytes(ctx, reswire.InternalizeMeta(in.Target))
	return &reswire.GetStatusOut{
		Exists: exists,
		Status: data,
	}, err
}

func (d *ResourceService) UpdateStatus(ctx context.Context, in *reswire.UpdateStatusIn) (*reswire.UpdateStatusOut, error) {
	which := internalizeMeta(in.Target)
	exists, err := d.client.UpdateStatusBytes(ctx, which, in.Status)
	return &reswire.UpdateStatusOut{Exists: exists}, err
}

func (d *ResourceService) GetEvents(ctx context.Context, in *reswire.GetEventsIn) (*reswire.GetEventsOut, error) {
	events, exists, err := d.client.GetLogs(ctx, internalizeMeta(in.Ref), reswire.InternalizeEventLevel(in.Level))
	if err != nil {
		return nil, err
	}

	out := &reswire.GetEventsOut{
		Exists: exists,
	}
	out.Events = make([]*reswire.Event, len(events))
	for i, e := range events {
		out.Events[i] = &reswire.Event{
			When:     timestamppb.New(e.When),
			Level:    reswire.ExternalizeEventLevel(e.Level),
			Rendered: fmt.Sprintf(e.Format, e.Params...),
		}
	}
	return out, nil
}

func (d *ResourceService) Log(ctx context.Context, in *reswire.LogIn) (*reswire.LogOut, error) {
	which := internalizeMeta(in.Ref)
	level := reswire.InternalizeEventLevel(in.Event.Level)
	//when := in.Event.When.AsTime()
	msg := in.Event.Rendered

	exists, err := d.client.Log(ctx, which, level, msg)
	if err != nil {
		return nil, err
	}
	return &reswire.LogOut{Exists: exists}, nil
}

func (d *ResourceService) List(ctx context.Context, in *reswire.ListIn) (*reswire.ListOut, error) {
	kind := reswire.InternalizeKind(in.Type)
	refs, err := d.client.List(ctx, kind)
	if err != nil {
		return nil, err
	}
	result := &reswire.ListOut{
		Ref: make([]*reswire.Meta, len(refs)),
	}
	for i, r := range refs {
		result.Ref[i] = reswire.MetaToWire(r)
	}
	return result, nil
}

func internalizeMeta(meta *reswire.Meta) resources.Meta {
	return resources.Meta{
		Type: reswire.InternalizeKind(meta.Kind),
		Name: meta.Name,
	}
}

func (d *ResourceService) Watcher(w reswire.ResourceController_WatcherServer) error {
	ctx := w.Context()
	watcher, err := d.client.Watcher(ctx)
	if err != nil {
		return err
	}

	events := make(chan *reswire.WatcherEventIn, 4)
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
		}
	}
	return result
}
