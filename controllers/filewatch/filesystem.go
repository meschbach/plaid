package filewatch

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"time"
)

type FileSystem interface {
	Watch(ctx context.Context, path string) error
	ChangeFeed() <-chan ChangeEvent
}

type ChangeKind uint8

const (
	FileModified ChangeKind = iota
)

type ChangeEvent struct {
	Kind    ChangeKind
	Path    string
	When    time.Time
	Tracing trace.SpanContext
}
