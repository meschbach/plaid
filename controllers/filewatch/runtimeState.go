package filewatch

import (
	"context"
	"errors"
	"github.com/meschbach/plaid/resources"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"strings"
)

// runtimeState is runtime data specific to a specific Controller instance.
type runtimeState struct {
	resources *resources.Client
	fs        FileSystem
	watchers  []*watch
}

func (r *runtimeState) registerWatcher(ctx context.Context, path string, who *watch) error {
	r.watchers = append(r.watchers, who)
	return nil
}

func (r *runtimeState) consumeFSEvent(parent context.Context, event ChangeEvent) error {
	ctx, span := tracing.Start(parent, "FileWatch/runtime.consumeFSEvent", trace.WithAttributes(attribute.String("file.Path", event.Path)))
	defer span.End()

	var problems []error
	for _, w := range r.watchers {
		if strings.HasPrefix(event.Path, w.base) {
			w.lastUpdated = &event.When
			if err := w.flushStatus(ctx); err != nil {
				problems = append(problems, err)
			}
		}
	}
	return errors.Join(problems...)
}
