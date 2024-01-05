package daemon

import (
	"context"
	"errors"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/daemon/wire"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"github.com/thejerf/suture/v4"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"sync/atomic"
)

type wireAdapterHandler struct {
	handler    resources.OnResourceChanged
	parentSpan trace.SpanContext
}

// watcherAdapter is a client to fulfill watcher semantics.
type watcherAdapter struct {
	wire   wire.ResourceControllerClient
	stream wire.ResourceController_WatcherClient
	tags   map[resources.WatchToken]*wireAdapterHandler
	next   atomic.Int32
}

func (w *watcherAdapter) Serve(ctx context.Context) error {
	for {
		e, err := w.stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
				return suture.ErrDoNotRestart
			} else {
				if s, ok := status.FromError(err); ok {
					if s.Code() == codes.Canceled {
						if err := w.stream.CloseSend(); err != nil {
							return errors.Join(err, suture.ErrDoNotRestart)
						}
						return suture.ErrDoNotRestart
					}
				}
				return err
			}
		}
		if err := w.dispatch(ctx, e); err != nil {
			return err
		}
	}
}

func (w *watcherAdapter) dispatch(serviceContext context.Context, e *wire.WatcherEventOut) error {
	if op, has := w.tags[resources.WatchToken(e.Tag)]; has {
		operation := internalizeOperation(e.Op)
		which := internalizeMeta(e.Ref)

		parentContext := trace.ContextWithRemoteSpanContext(serviceContext, op.parentSpan)
		ctx, span := tracer.Start(parentContext, operation.String()+" of "+which.Type.String())
		defer span.End()
		if err := op.handler(ctx, resources.ResourceChanged{
			Which:     which,
			Operation: operation,
			Tracing:   trace.LinkFromContext(ctx),
		}); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
				span.AddEvent("eof or canceled")
				return suture.ErrDoNotRestart
			} else {
				span.SetStatus(otelcodes.Error, "unexpected wire error")
				return err
			}
		}
	} else {
		//todo: resource went missing?
	}
	return nil
}

func (w *watcherAdapter) OnType(ctx context.Context, kind resources.Type, consume resources.OnResourceChanged) (resources.WatchToken, error) {
	tag := w.next.Add(1)
	k := typeToWire(kind)
	if err := w.stream.Send(&wire.WatcherEventIn{
		Tag:    uint64(tag),
		OnType: k,
	}); err != nil {
		return 0, err
	}
	token := resources.WatchToken(tag)
	w.tags[token] = &wireAdapterHandler{handler: consume, parentSpan: trace.SpanContextFromContext(ctx)}
	return token, nil
}

func (w *watcherAdapter) OnResource(parent context.Context, ref resources.Meta, consume resources.OnResourceChanged) (resources.WatchToken, error) {
	_, span := tracer.Start(parent, "Watcher.OnResource")
	defer span.End()
	tag := w.next.Add(1)
	wireRef := metaToWire(ref)
	if err := w.stream.Send(&wire.WatcherEventIn{
		Tag:        uint64(tag),
		OnResource: wireRef,
	}); err != nil {
		return 0, err
	}
	token := resources.WatchToken(tag)
	span.SetAttributes(attribute.Int64("tag", int64(tag)), attribute.Int64("token", int64(token)))
	w.tags[token] = &wireAdapterHandler{
		handler:    consume,
		parentSpan: trace.SpanContextFromContext(parent),
	}
	return token, nil
}

func noOpChangeListener(ctx context.Context, changed resources.ResourceChanged) error {
	return nil
}

func (w *watcherAdapter) Off(ctx context.Context, token resources.WatchToken) error {
	w.tags[token].handler = noOpChangeListener
	t := true
	return w.stream.Send(&wire.WatcherEventIn{
		Tag:    uint64(token),
		Delete: &t,
	})
}

func (w *watcherAdapter) Close(ctx context.Context) error {
	return w.stream.CloseSend()
}
