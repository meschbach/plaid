package filewatch

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/trace"
	"time"
)

type FileSystem interface {
	Watch(ctx context.Context, path string) error
	Unwatch(ctx context.Context, path string) error
	ChangeFeed() <-chan ChangeEvent
}

type ChangeKind uint8

const (
	FileModified ChangeKind = iota
)

func (c ChangeKind) String() string {
	switch c {
	case FileModified:
		return "File Modified"
	default:
		panic(fmt.Sprintf("Unknown change kind %#v\n", c))
	}
}

type ChangeEvent struct {
	Kind    ChangeKind
	Path    string
	When    time.Time
	Tracing trace.SpanContext
}

func (c ChangeEvent) String() string {
	return fmt.Sprintf("filewatch.ChangeEvent{ %s @ %s [%s] }", c.Kind, c.Path, c.When.Format("2006-01-02T15:04:05"))
}
