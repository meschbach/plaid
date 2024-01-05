package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"testing"
)

func AssertStatus[T any](t *testing.T, parent context.Context, res *Client, ref Meta, check func(status T) bool) {
	t.Helper()

	ctx, span := tracing.Start(parent, "AssertStatus")
	defer span.End()

	w, err := res.Watcher(ctx)
	if !assert.NoError(t, err) {
		return
	}
	matched := false
	if !assert.NoError(t, w.WatchResource(ctx, ref, func(ctx context.Context, changed ResourceChanged) error {
		ctx, span := tracing.Start(parent, "AssertStatus#check")
		defer span.End()

		var status T
		exists, err := res.GetStatus(ctx, changed.Which, &status)
		require.NoError(t, err)
		if !exists {
			span.AddEvent("does-not-exist")
			return nil
		}

		span.SetAttributes(attribute.String("resource", fmt.Sprintf("%#v", changed)))
		matched = check(status)
		return nil
	})) {
		return
	}

	//initial check
	var status T
	exists, err := res.GetStatus(ctx, ref, &status)
	require.NoError(t, err)
	if exists {
		matched = check(status)
		span.AddEvent("existing check", trace.WithAttributes(attribute.Bool("matched", matched)))
	}

	for !matched {
		select {
		case e := <-w.Feed:
			span.AddEvent("digesting event")
			require.NoError(t, w.Digest(ctx, e))
		case <-ctx.Done():
			fmt.Printf("timed out waiting for %#v to match expectations", status)
			span.SetStatus(codes.Error, "timed out")
			if !assert.NoError(t, ctx.Err()) {
				return
			}
		}
	}
	if !matched {
		span.SetStatus(codes.Error, "not matched")
	}
	require.True(t, matched, "matched expectations")
}

func RequireCreate(t *testing.T, ctx context.Context, res *Client, kind Type, name string, spec any) Meta {
	ref := Meta{
		Type: kind,
		Name: name,
	}
	assert.NoError(t, res.Create(ctx, ref, spec))
	return ref
}

func AssertExists(t *testing.T, ctx context.Context, res *Client, ref Meta) {
	t.Helper()
	op := make(chan interface{}, 16)
	w, err := res.Watcher(ctx)
	assert.NoError(t, err)
	_, err = w.OnResource(ctx, ref, func(ctx context.Context, changed ResourceChanged) error {
		op <- nil
		return nil
	})
	assert.NoError(t, err)
	op <- nil

	for {
		select {
		case <-op:
			var payload json.RawMessage
			exists, err := res.Get(ctx, ref, &payload)
			assert.NoError(t, err)
			if exists {
				return
			}
		case <-ctx.Done():
			assert.NoError(t, ctx.Err())
		}
	}
}
