package local

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/controllers/logdrain"
	"github.com/meschbach/plaid/internal/plaid/resources"
	"github.com/meschbach/go-junk-bucket/pkg/reactors"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
	"github.com/meschbach/go-junk-bucket/pkg/task"
	"go.opentelemetry.io/otel/trace"
	"time"
)

type logRelay struct {
	config    *logdrain.ServiceConfig
	from      <-chan string
	logBuffer *streams.Buffer[logdrain.LogEntry]

	ref        resources.Meta
	streamName string
	fromSpan   trace.SpanContext
}

func (l *logRelay) Serve(ctx context.Context) error {
	loop, loopIn := reactors.NewChannel[*logRelay](128)

	setup := false
	logging := logdrain.BuildClient[*logRelay](ctx, l.config, loop)
	loop.ScheduleStateFunc(ctx, func(parent context.Context, state *logRelay) error {
		sourceContext := trace.ContextWithRemoteSpanContext(parent, l.fromSpan)
		ctx, span := tracer.Start(sourceContext, "localexec."+l.streamName+".Init")
		defer span.End()
		p := logging.RegisterSource(ctx, l.ref, l.streamName, l.logBuffer)

		running := true
		var result *task.Result[*logdrain.SourceConnection[*logRelay]]
		p.OnCompleted(ctx, func(ctx context.Context, event task.Result[*logdrain.SourceConnection[*logRelay]]) {
			result = &event
			running = false
		})
		for running {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case e := <-loopIn:
				if err := loop.Tick(ctx, e, state); err != nil {
					return err
				}
			}
		}
		setup = true
		return result.Problem
	})
	for !setup {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-loopIn:
			if err := loop.Tick(ctx, e, l); err != nil {
				return err
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-loopIn:
			if err := loop.Tick(ctx, e, l); err != nil {
				return err
			}
		case line, ok := <-l.from:
			loop.ScheduleFunc(ctx, func(parent context.Context) error {
				sourceContext := trace.ContextWithRemoteSpanContext(parent, l.fromSpan)

				if ok {
					ctx, span := tracer.Start(sourceContext, "localexec."+l.streamName+".LogLine")
					defer span.End()
					if err := l.logBuffer.Write(ctx, logdrain.LogEntry{
						When:    time.Now(),
						Message: line,
					}); err != nil {
						return err
					}
				} else {
					_, span := tracer.Start(sourceContext, "localexec."+l.streamName+".Finish")
					defer span.End()
					if err := l.logBuffer.Finish(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		}
	}
}
