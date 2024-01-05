package resources

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Metadata struct {
	For         Meta
	Annotations map[string]string
	ClaimedBy   []Meta
}

func (c *Client) GetMetadataFor(ctx context.Context, which Meta) (*Metadata, bool, error) {
	replyTo := make(chan getMetadataReply, 1)

	select {
	case c.dataPlane <- &getMetadataOp{
		replyTo:       replyTo,
		ref:           which,
		parentContext: trace.SpanContextFromContext(ctx),
	}:
	case <-ctx.Done():
		return nil, false, ctx.Err()
	}

	select {
	case <-ctx.Done():
		return nil, false, ctx.Err()
	case reply := <-replyTo:
		return &reply.data, reply.found, reply.problem
	}
}

type getMetadataReply struct {
	found   bool
	problem error
	data    Metadata
}

type getMetadataOp struct {
	replyTo       chan<- getMetadataReply
	ref           Meta
	parentContext trace.SpanContext
}

func (g *getMetadataOp) perform(serviceContext context.Context, c *Controller) {
	defer close(g.replyTo)
	reply := func(reply getMetadataReply) {
		select {
		case <-serviceContext.Done():
		case g.replyTo <- reply:
		}
	}
	_, span := TracedMessageContext(serviceContext, g.parentContext, "resources.GetMetadataFor "+g.ref.Type.String(), trace.WithAttributes(attribute.String("name", g.ref.Name)))
	defer span.End()

	node, exists := c.resources.Find(g.ref)
	if !exists {
		reply(getMetadataReply{
			found:   false,
			problem: nil,
		})
	}

	reply(getMetadataReply{
		found:   true,
		problem: nil,
		data: Metadata{
			For:         g.ref,
			Annotations: node.annotations,
			ClaimedBy:   node.claimedBy,
		},
	})
}
