package resources

import (
	"bytes"
	"context"
	"errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type createReply struct {
	problem error
}

type createOp struct {
	meta          Meta
	spec          []byte
	replyTo       chan<- createReply
	parentContext trace.SpanContext
	claims        []Meta
	annotations   map[string]string
	closeChannel  bool
}

func (o *createOp) perform(serviceContext context.Context, c *Controller) {
	ctx, span := TracedMessageContext(serviceContext, o.parentContext, "resources.Create "+o.meta.Type.String(), trace.WithAttributes(attribute.String("name", o.meta.Name)))
	defer span.End()

	//todo: rpc -- create context linked between both controller actor and calling context
	node, created, err := c.createOrGetNode(ctx, o.meta)
	if err != nil {
		o.replyTo <- createReply{problem: err}
		return
	}
	if created {
		node.spec = o.spec
		node.claimedBy = o.claims
		node.annotations = o.annotations
		c.dispatchUpdates(ctx, CreatedEvent, o.meta)
	} else {
		//todo: compare the two values to see if there is a legitimate change
		if bytes.Equal(node.spec, o.spec) {
			//do nothing since they are
		} else {
			o.replyTo <- createReply{problem: errors.New("already exists")}
			return
		}
	}
	//creation done successfully
	//todo: prevent hanging when client is not receiving
	o.replyTo <- createReply{problem: nil}
	if o.closeChannel {
		close(o.replyTo)
	}
}

func WithAnnotations(annotations map[string]string) CreateOpt {
	return createOptFunc(func(op *createOp) {
		if op.annotations != nil {
			panic("multiple annotations")
		}
		op.annotations = annotations
	})
}
