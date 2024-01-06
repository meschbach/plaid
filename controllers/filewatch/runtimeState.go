package filewatch

import (
	"context"
	"github.com/meschbach/plaid/resources"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"strings"
)

// runtimeState is runtime data specific to a specific Controller instance.
type runtimeState struct {
	resources *resources.Client
	fs        filesystem
	watchers  []*watch
}

func (r *runtimeState) registerWatcher(ctx context.Context, path string, who *watch) error {
	if who.recursive {
		var step func(string) error
		step = func(e string) error {
			if err := r.fs.Watch(ctx, e); err != nil { //todo: should probably log and skip it
				return err
			}
			dirs, err := r.fs.ListDirectories(e)
			if err != nil { //todo: should probably just skip over it and log it
				return err
			}

			for _, d := range dirs {
				if err := step(d); err != nil {
					return err
				}
			}
			return nil
		}
		if err := step(path); err != nil {
			return err
		}
	} else {
		if err := r.fs.Watch(ctx, path); err != nil {
			return err
		}
	}
	r.watchers = append(r.watchers, who)
	return nil
}

func (r *runtimeState) consumeFSEvent(parent context.Context, event changeEvent) error {
	ctx, span := tracing.Start(parent, "FileWatch/runtime.consumeFSEvent", trace.WithAttributes(attribute.String("file.path", event.path)))
	defer span.End()

	for _, w := range r.watchers {
		if w.recursive {
			if strings.HasPrefix(event.path, w.base) {
				span.AddEvent("recursive-match", trace.WithAttributes(attribute.String("prefix", w.base), attribute.String("path", event.path)))
				if err := w.updateStatus.fileChanged(ctx, w.meta, r, event); err != nil {
					return err
				}
			}
		} else {
			if event.path == w.base {
				span.AddEvent("match", trace.WithAttributes(attribute.String("path", w.base), attribute.String("path", event.path)))
				if err := w.updateStatus.fileChanged(ctx, w.meta, r, event); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
