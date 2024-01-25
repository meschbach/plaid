package fsn

import (
	"context"
	"github.com/fsnotify/fsnotify"
	"github.com/meschbach/plaid/controllers/filewatch"
	"go.opentelemetry.io/otel/trace"
	"time"
)

func (c *Core) consumeFSEvent(ctx context.Context, e fsnotify.Event) error {
	traceContext := trace.SpanContextFromContext(ctx)
	if e.Has(fsnotify.Write) || e.Has(fsnotify.Create) {
		select {
		case c.Output <- filewatch.ChangeEvent{
			Kind:    filewatch.FileModified,
			Path:    e.Name,
			When:    time.Now(),
			Tracing: traceContext,
		}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (c *Core) ChangeFeed() <-chan filewatch.ChangeEvent {
	return c.Output
}
