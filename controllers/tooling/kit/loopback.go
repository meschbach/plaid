package kit

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/resources"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type loopbackOp uint8

const (
	loopbackUpdateStatus loopbackOp = iota
)

type LoopbackEvent struct {
	op       loopbackOp
	ref      resources.Meta
	causedBy trace.Link
}

func (k *Kit[Spec, Status, State]) DigestLoopback(parent context.Context, event LoopbackEvent) error {
	switch event.op {
	case loopbackUpdateStatus:
		ctx, span := tracer.Start(parent, "Kit["+k.kind.String()+"].Loopback/UpdateStatus",
			trace.WithAttributes(event.ref.AsTraceAttribute("ref")...),
			trace.WithLinks(event.causedBy),
			trace.WithSpanKind(trace.SpanKindConsumer),
		)
		defer span.End()

		state, has := k.mapping.Find(event.ref)
		if !has { //deleted in between?
			span.AddEvent("missing")
			return nil
		}

		err := k.updateStatus(ctx, span, event.ref, state)
		if err != nil {
			span.SetStatus(codes.Error, "failed to update status")
		}
		return err
	default:
		return fmt.Errorf("unknown opreation %d\n", event.op)
	}
}

type loopbackManager struct {
	target chan<- LoopbackEvent
	ref    resources.Meta
}

func (l *loopbackManager) UpdateStatus(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case l.target <- LoopbackEvent{
		op:       loopbackUpdateStatus,
		ref:      l.ref,
		causedBy: trace.LinkFromContext(ctx),
	}:
		return nil
	}
}
