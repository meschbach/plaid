package logdrain

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/resources"
	"github.com/meschbach/go-junk-bucket/pkg/reactors"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
	"github.com/meschbach/go-junk-bucket/pkg/task"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"sync"
	"time"
)

type ChangedEvent struct {
	Resource resources.Meta
	Name     string
}

type LogEntryOrigin struct {
	From   resources.Meta
	Stream string
}

type LogEntry struct {
	Origin  LogEntryOrigin
	When    time.Time
	Message string
}

// todo: should not go through the logddrain reactor
type core struct {
	output *streams.FanOutSink[LogEntry]
}

type Client[T any] struct {
	clientWell reactors.Boundary[T]
	//logDrainWell targets the log drain's looping system
	logDrainWell reactors.Boundary[*core]
}

func failedPromise[T any](ctx context.Context, problem error) *task.Promise[T] {
	p := &task.Promise[T]{}
	p.Failure(ctx, problem)
	return p
}

func (c *Client[T]) RegisterDrain(ctx context.Context, sinkTarget streams.Sink[LogEntry]) *task.Promise[bool] {
	clientOutflow, logDrainInflow, err := reactors.StreamBetween[LogEntry, *core, T](ctx, c.logDrainWell, c.clientWell, reactors.WithStreamBetweenName("drain"))
	if err != nil {
		return failedPromise[bool](ctx, err)
	}

	barrier := &sync.WaitGroup{}
	barrier.Add(1)
	c.clientWell.ScheduleFunc(ctx, func(ctx context.Context) error {
		//todo: handle connected state for de-registration
		_, err = streams.Connect[LogEntry](ctx, clientOutflow, sinkTarget, streams.WithTracePrefix("log.drain"))
		barrier.Done()
		return err
	})

	return reactors.Submit(ctx, c.clientWell, c.logDrainWell, func(boundaryContext context.Context, state *core) (bool, error) {
		reactors.VerifyWithinBoundary(boundaryContext, c.logDrainWell)
		state.output.Add(logDrainInflow)
		barrier.Wait()
		return true, nil
	})
}

type SourceConnection[T any] struct {
}

func traceStreamAttributes(attributePrefix string, ref resources.Meta, streamName string) trace.SpanStartEventOption {
	return trace.WithAttributes(attribute.Stringer(attributePrefix+".ref", ref), attribute.String(attributePrefix+".stream", streamName))
}

func (c *Client[T]) RegisterSource(parent context.Context, ref resources.Meta, name string, from streams.Source[LogEntry]) *task.Promise[*SourceConnection[T]] {
	ctx, span := tracing.Start(parent, "logdrain.RegisterSource", traceStreamAttributes("source", ref, name))
	defer span.End()
	reactors.VerifyWithinBoundary[T](ctx, c.clientWell)

	logdrainChannelOutput, clientChannelInput, err := reactors.StreamBetween[LogEntry, *core, T](ctx, c.logDrainWell, c.clientWell, reactors.WithStreamBetweenName("log.source"))
	if err != nil {
		return failedPromise[*SourceConnection[T]](ctx, err)
	}

	return reactors.Submit(ctx, c.clientWell, c.logDrainWell, func(boundaryContext context.Context, state *core) (*SourceConnection[T], error) {
		ctx, span := tracing.Start(boundaryContext, "logdrain.RegisterSource#logDrainWell", traceStreamAttributes("source", ref, name))
		defer span.End()

		wrapper := streams.WrapTransformingSink[LogEntry, LogEntry](state.output, func(ctx context.Context, in LogEntry) (LogEntry, error) {
			in.Origin.From = ref
			in.Origin.Stream = name
			return in, nil
		})
		_, err := streams.Connect(ctx, logdrainChannelOutput, wrapper, streams.WithTracePrefix(ref.String()+"#logdrain"), streams.WithSuppressEnd())
		if err != nil {
			return nil, err
		}
		gate := sync.WaitGroup{}
		gate.Add(1)
		reactors.Submit(ctx, c.logDrainWell, c.clientWell, func(boundaryContext context.Context, state T) (*SourceConnection[T], error) {
			defer gate.Done()
			ctx, span := tracing.Start(parent, "logdrain.RegisterSource#clientWell", traceStreamAttributes("source", ref, name))
			defer span.End()
			_, err := streams.Connect(ctx, from, clientChannelInput, streams.WithTracePrefix(ref.String()+"#producer"))
			return nil, err
		})
		gate.Wait()
		return nil, nil
	})
}
