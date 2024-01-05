package resources

import (
	"context"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type getStatusReply struct {
	status  []byte
	exists  bool
	problem error
}

type getStatusOp struct {
	meta    Meta
	replyTo chan<- getStatusReply
	tracing trace.SpanContext
}

func (g *getStatusOp) name() string {
	return "get-status"
}

func (g *getStatusOp) perform(parent context.Context, c *Controller) {
	ctx, span := TracedMessageContext(parent, g.tracing, "Controller.GetStatus")
	defer span.End()

	defer close(g.replyTo)
	doReply := func(r getStatusReply) {
		select {
		case g.replyTo <- r:
		case <-ctx.Done():
		}
	}

	node, err := c.getNode(ctx, g.meta)
	if err != nil {
		span.SetStatus(codes.Error, "getNode")
		span.RecordError(err)
		doReply(getStatusReply{problem: err})
		return
	}
	if node == nil {
		doReply(getStatusReply{exists: false})
		return
	}
	doReply(getStatusReply{
		status:  node.status,
		exists:  true,
		problem: nil,
	})
}
