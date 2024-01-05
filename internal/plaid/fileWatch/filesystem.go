package fileWatch

import (
	"context"
	"go.opentelemetry.io/otel/trace"
)

type filesystem interface {
	//ListDirectories recursively lists directories from the given base path
	ListDirectories(path string) ([]string, error)
	Watch(ctx context.Context, path string) error
	ChangeFeed() <-chan changeEvent
}

type fsChangeEventKind uint8

const (
	fsFileModified fsChangeEventKind = iota
)

type changeEvent struct {
	path    string
	kind    fsChangeEventKind
	tracing trace.SpanContext
}
