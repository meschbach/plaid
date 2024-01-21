package resources

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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

func (c *Client) Create(parent context.Context, meta Meta, resource any, opts ...CreateOpt) error {
	jsonEncodedResource, err := json.Marshal(resource)
	if err != nil {
		return err
	}

	return c.CreateBytes(parent, meta, jsonEncodedResource, opts...)
}

type CreateOpt interface {
	apply(op *createOp)
}

type createOptFuncWrapper struct {
	f func(op *createOp)
}

func (c *createOptFuncWrapper) apply(o *createOp) {
	c.f(o)
}

func createOptFunc(f func(op *createOp)) CreateOpt {
	return &createOptFuncWrapper{f}
}

func (c *Client) CreateBytes(ctx context.Context, meta Meta, resource []byte, opts ...CreateOpt) error {
	span := trace.SpanFromContext(ctx)
	resultSignal, has := ctx.Value(signalCreateChannel).(chan createReply)
	closeChannel := false
	if !has {
		resultSignal = make(chan createReply)
		closeChannel = true
	}
	op := &createOp{
		meta:          meta,
		spec:          resource,
		replyTo:       resultSignal,
		parentContext: trace.SpanContextFromContext(ctx),
		closeChannel:  closeChannel,
	}
	for _, opt := range opts {
		opt.apply(op)
	}

	select {
	case c.dataPlane <- op:
	case <-ctx.Done():
		span.SetStatus(codes.Error, "context done before create sent")
		return ctx.Err()
	}
	select {
	case result := <-resultSignal:
		if result.problem != nil {
			span.SetStatus(codes.Error, result.problem.Error())
			span.RecordError(result.problem)
		}
		return result.problem
	case <-ctx.Done():
		span.SetStatus(codes.Error, "context done before create finished")
		return ctx.Err()
	}
}

const signalCreateChannel = "plaid.resource.create"

func WithFastCreate(parent context.Context) context.Context {
	create := make(chan createReply)
	go func() {
		<-parent.Done()
		close(create)
	}()
	return context.WithValue(parent, signalCreateChannel, create)
}
