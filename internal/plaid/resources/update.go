package resources

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type updateReply struct {
	problem error
	exists  bool
}

type updateOp struct {
	replyTo chan updateReply
	what    Meta
	spec    []byte
	tracing trace.SpanContext
}

func (u *updateOp) name() string {
	return "update"
}

func (u *updateOp) perform(parent context.Context, c *Controller) {
	ctx, span := TracedMessageContext(parent, u.tracing, "resources.Create "+u.what.Type.String(), trace.WithAttributes(attribute.String("name", u.what.Name)))
	defer span.End()

	defer close(u.replyTo)
	reply := func(exists bool, problem error) {
		if problem != nil {
			span.SetStatus(codes.Error, "failed")
		}
		select {
		case u.replyTo <- updateReply{
			problem: problem,
			exists:  exists,
		}:
		case <-ctx.Done():
		}
	}

	node, err := c.getNode(ctx, u.what)
	if err != nil {
		reply(false, err)
		return
	}
	if node == nil {
		reply(false, nil)
		return
	}

	//update
	node.spec = u.spec

	//dispatch update notices
	c.dispatchUpdates(ctx, UpdatedEvent, u.what)

	reply(true, nil)
}
