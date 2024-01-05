package junk

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/observability"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"testing"
	"time"
)

var setup = false

func SetupTestTracing(t *testing.T) func(context.Context) {
	if setup {
		return func(ctx context.Context) {}
	}

	cfg := observability.DefaultConfig("plaid")
	cfg.Silent = true
	cfg.Environment = "test"
	cfg.Batched = false
	setupContext, done := context.WithTimeout(context.Background(), 10*time.Second)
	defer done()
	c, err := cfg.Start(setupContext)
	require.NoError(t, err)
	setup = true

	return func(shutdownContext context.Context) {
		if err := c.ShutdownGracefully(shutdownContext); err != nil {
			panic(err)
		}
	}
}

func TraceSubtest(t *testing.T, parentContext context.Context, tracer trace.Tracer) (context.Context, trace.Span) {
	ctx, span := tracer.Start(parentContext, t.Name())
	t.Cleanup(func() {
		if t.Failed() {
			span.SetStatus(codes.Error, "test failed")
		}
		span.End()
	})
	return ctx, span
}
