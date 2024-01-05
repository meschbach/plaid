package logger

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
)

type loggingFacet interface {
	registerDrain(ctx context.Context, name string, output streams.Sink[bufferedEntry]) (bool, error)
	unregisterDrain(ctx context.Context, name string) error
}
