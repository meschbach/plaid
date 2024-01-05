package resources

import (
	"bytes"
	"context"
	"encoding/json"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type updateStatusReply struct {
	problem error
	exists  bool
}

type updateStatusOp struct {
	replyTo chan updateStatusReply
	what    Meta
	status  []byte
	tracing trace.SpanContext
}

func (c *Client) UpdateStatus(ctx context.Context, what Meta, status any) (bool, error) {
	payload, err := json.Marshal(status)
	if err != nil {
		return false, err
	}

	return c.UpdateStatusBytes(ctx, what, payload)
}

// UpdateStatusBytes changes the status of the specified resource.  Returns existence and error
func (c *Client) UpdateStatusBytes(ctx context.Context, what Meta, status []byte) (bool, error) {
	resultSignal := make(chan updateStatusReply, 1)
	//todo -- op should close the channel
	select {
	case c.dataPlane <- &updateStatusOp{
		replyTo: resultSignal,
		what:    what,
		status:  status,
		tracing: trace.SpanContextFromContext(ctx),
	}:
		select {
		case msg := <-resultSignal:
			return msg.exists, msg.problem
		case <-ctx.Done():
			return false, ctx.Err()
		}
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

func (u *updateStatusOp) perform(serviceContext context.Context, c *Controller) {
	ctx, span := TracedMessageContext(serviceContext, u.tracing, "resources.UpdateStatus of "+u.what.Type.String(), trace.WithAttributes(attribute.String("name", u.what.Name)))
	defer span.End()

	defer close(u.replyTo)

	reply := func(reply updateStatusReply) {
		select {
		case u.replyTo <- reply:
		case <-ctx.Done():
		}
	}

	node, err := c.getNode(ctx, u.what)
	if err != nil {
		reply(updateStatusReply{
			problem: err,
			exists:  false,
		})
		return
	}
	if node == nil {
		reply(updateStatusReply{
			problem: nil,
			exists:  false,
		})
	}

	changed := !bytes.Equal(node.status, u.status)
	if changed {
		node.status = u.status
		c.dispatchUpdates(ctx, StatusUpdated, u.what)
	}

	reply(updateStatusReply{
		problem: nil,
		exists:  true,
	})
}
