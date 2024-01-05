package resources

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type claimsReply struct {
	exists  bool
	problem error
}

type claimsOp struct {
	owned   Meta
	owner   Meta
	replyTo chan<- claimsReply
	tracing trace.SpanContext
}

func (o *claimsOp) name() string {
	return "owns"
}

func (o *claimsOp) perform(parent context.Context, c *Controller) {
	ctx, span := TracedMessageContext(parent, o.tracing, "Controller.Claims")
	defer span.End()

	doReply := func(r claimsReply) {
		select {
		case o.replyTo <- r:
		case <-ctx.Done():
		}
	}

	node, err := c.getNode(ctx, o.owned)
	if err != nil {
		span.SetStatus(codes.Error, "getNode")
		span.RecordError(err)
		doReply(claimsReply{problem: err})
		return
	}
	if node == nil {
		doReply(claimsReply{exists: false})
		return
	}
	node.claimedBy = append(node.claimedBy, o.owner)
	doReply(claimsReply{
		exists:  true,
		problem: nil,
	})
}

// Claims notes the claimed resource is managed by the claimer in some reasonable way
func (c *Client) Claims(ctx context.Context, claimed, claimer Meta) (exists bool, problem error) {
	resultSignal := make(chan claimsReply, 1)
	select {
	case c.dataPlane <- &claimsOp{
		owned:   claimed,
		owner:   claimer,
		replyTo: resultSignal,
		tracing: trace.SpanContextFromContext(ctx),
	}:
	case <-ctx.Done():
		return false, ctx.Err()
	}
	select {
	case result := <-resultSignal:
		return result.exists, result.problem
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

type findClaimedBy struct {
	found   []Meta
	problem error
}

type findClaimedByOp struct {
	owner   Meta
	types   []Type
	replyTo chan<- findClaimedBy
	tracing trace.SpanContext
}

func (f *findClaimedByOp) perform(parent context.Context, c *Controller) {
	ctx, span := TracedMessageContext(parent, f.tracing, "Controller.FindClaimedBy")
	defer span.End()

	doReply := func(r findClaimedBy) {
		select {
		case f.replyTo <- r:
		case <-ctx.Done():
		}
	}

	if f.types == nil {
		f.types = c.resources.AllTypes()
	}
	span.SetAttributes(attribute.Int("types.count", len(f.types)))

	var matching []Meta
	for _, resourceType := range f.types {
		names := c.resources.ListNames(resourceType)
		for _, name := range names {
			ref := Meta{
				Type: resourceType,
				Name: name,
			}
			node, has := c.resources.Find(ref)
			//strange case, I do not think this would ever happen...which are also famous last words
			if !has {
				span.AddEvent("missing", trace.WithAttributes(attribute.Stringer("ref", ref)))
				continue
			}

			if MetaSliceContains(node.claimedBy, f.owner) {
				matching = append(matching, ref)
			}
		}
	}
	span.SetAttributes(attribute.Int("found", len(matching)))
	doReply(findClaimedBy{
		found:   matching,
		problem: nil,
	})
}

// FindClaimedBy searches through the given types to find the resources claimed by the given resource
func (c *Client) FindClaimedBy(ctx context.Context, claimer Meta, types []Type) ([]Meta, error) {
	resultSignal := make(chan findClaimedBy, 1)
	select {
	case c.dataPlane <- &findClaimedByOp{
		owner:   claimer,
		replyTo: resultSignal,
		types:   types,
		tracing: trace.SpanContextFromContext(ctx),
	}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	select {
	case result := <-resultSignal:
		return result.found, result.problem
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func ClaimedBy(claimer Meta) CreateOpt {
	return createOptFunc(func(op *createOp) {
		op.claims = append(op.claims, claimer)
	})
}
