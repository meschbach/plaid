package logdrain

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/resources"
	"github.com/meschbach/plaid/internal/plaid/resources/operator"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type alpha1State struct {
	bridge *operator.KindBridgeState
	self   resources.Meta

	pipeConnection *streams.ConnectedPipe[LogEntry]
	source         streams.Source[LogEntry]
	drain          streams.Sink[LogEntry]
}

func (a *alpha1State) updateSource(ctx context.Context, source streams.Source[LogEntry]) error {
	if source == nil {
		a.source = nil
		return a.disconnect(ctx)
	}
	if source == a.source {
		return nil
	}
	a.source = source
	return a.maybeConnect(ctx)
}

func (a *alpha1State) updateDrain(ctx context.Context, drain streams.Sink[LogEntry]) error {
	if drain == nil {
		a.drain = nil
		return a.disconnect(ctx)
	}
	if drain == a.drain {
		return nil //no change
	}
	a.drain = drain
	return a.maybeConnect(ctx)
}

func (a *alpha1State) maybeConnect(ctx context.Context) error {
	span := trace.SpanFromContext(ctx)
	if a.drain == nil {
		span.AddEvent("no-drain")
		return nil
	}
	if a.source == nil {
		span.AddEvent("no-source")
		return nil
	}
	return a.connect(ctx)
}

func (a *alpha1State) connect(ctx context.Context) error {
	if err := a.disconnect(ctx); err != nil {
		return err
	}
	span := trace.SpanFromContext(ctx)

	pipe, err := streams.Connect(ctx, a.source, a.drain)
	if err != nil {
		span.SetStatus(codes.Error, "failed to connect source and destination")
		return err
	}
	span.AddEvent("connected")
	a.pipeConnection = pipe

	return a.bridge.OnResourceChange(ctx, a.self)
}

func (a *alpha1State) disconnect(ctx context.Context) error {
	if a.pipeConnection == nil {
		return nil
	}
	span := trace.SpanFromContext(ctx)
	span.AddEvent("disconnecting")
	return a.pipeConnection.Close(ctx)
}

func (a *alpha1State) toStatus() Alpha1Status {
	status := Alpha1Status{
		Pipe: Unknown,
	}
	if a.source == nil && a.drain == nil {
		status.Pipe = Unconnected
	} else {
		status.Pipe = Connected
	}
	return status
}
