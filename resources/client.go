package resources

import (
	"context"
	"encoding/json"
	"github.com/meschbach/go-junk-bucket/pkg/reactors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Client struct {
	dataPlane chan<- dataOp
	boundary  reactors.Boundary[*Controller]
}

// Delete attempts to delete teh named resource, return true if the resource exists otherwise false.
func (c *Client) Delete(ctx context.Context, what Meta) (bool, error) {
	resultSignal := make(chan deleteResult, 1)
	defer close(resultSignal)
	select {
	case c.dataPlane <- &deleteOp{
		replyTo: resultSignal,
		what:    what,
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

func (c *Client) Update(ctx context.Context, what Meta, resource any) (exists bool, problem error) {
	bytes, err := json.Marshal(resource)
	if err != nil {
		return false, err
	}

	return c.UpdateBytes(ctx, what, bytes)
}

// UpdateBytes changes the specified resource specification
func (c *Client) UpdateBytes(ctx context.Context, what Meta, resource []byte) (exists bool, problem error) {
	resultSignal := make(chan updateReply, 1)
	select {
	case c.dataPlane <- &updateOp{
		replyTo: resultSignal,
		what:    what,
		spec:    resource,
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

func (c *Client) Get(ctx context.Context, what Meta, out any) (exists bool, problem error) {
	bytes, exists, err := c.GetBytes(ctx, what)
	if err != nil || !exists {
		return exists, err
	}

	return true, json.Unmarshal(bytes, out)
}

// GetBytes retrieves the specification for the desired state for given resource
func (c *Client) GetBytes(ctx context.Context, meta Meta) (body []byte, exists bool, problem error) {
	resultSignal := make(chan getSpecReply, 1)
	select {
	case c.dataPlane <- &getSpecOp{
		meta:    meta,
		replyTo: resultSignal,
	}:
	case <-ctx.Done():
		return nil, false, ctx.Err()
	}
	select {
	case result := <-resultSignal:
		return result.spec, result.exists, result.problem
	case <-ctx.Done():
		return nil, false, ctx.Err()
	}
}

func (c *Client) GetStatus(parent context.Context, what Meta, out any) (bool, error) {
	ctx, span := tracing.Start(parent, "GetStatus", trace.WithAttributes(attribute.Stringer("what", what)))
	defer span.End()

	bytes, exists, err := c.GetStatusBytes(ctx, what)
	if err != nil {
		span.SetStatus(codes.Error, "rpc")
		return false, err
	}

	if !exists {
		return exists, err
	}

	if out == nil || len(bytes) == 0 {
		span.AddEvent("zero length result")
		return true, nil
	}

	err = json.Unmarshal(bytes, out)
	if err != nil {
		span.SetStatus(codes.Error, "unmarshall")
	}
	return true, err
}

// GetStatusBytes retrieves the status of the given resource
func (c *Client) GetStatusBytes(ctx context.Context, meta Meta) (body []byte, exists bool, problem error) {
	resultSignal := make(chan getStatusReply, 1)
	select {
	case c.dataPlane <- &getStatusOp{
		meta:    meta,
		replyTo: resultSignal,
		tracing: trace.SpanContextFromContext(ctx),
	}:
	case <-ctx.Done():
		return nil, false, ctx.Err()
	}
	select {
	case result := <-resultSignal:
		return result.status, result.exists, result.problem
	case <-ctx.Done():
		return nil, false, ctx.Err()
	}
}

// Watch registers a new channel to receive events about a specific kind.
// Deprecated: Use Watcher with OnType
func (c *Client) Watch(ctx context.Context, kind Type) (chan ResourceChanged, error) {
	startReply := make(chan watchStartedReply, 1)
	defer close(startReply)
	out := make(chan ResourceChanged, 32)

	select {
	case c.dataPlane <- &watchStartOp{
		feedContext: ctx,
		feedChannel: out,
		what:        kind,
		started:     startReply,
	}:
	case <-ctx.Done():
		close(out)
		return nil, ctx.Err()
	}

	select {
	case reply := <-startReply:
		if reply.problem != nil {
			close(out)
			return nil, reply.problem
		}
		return out, nil
	case <-ctx.Done():
		close(out)
		return nil, ctx.Err()
	}
}

func (c *Client) Watcher(ctx context.Context) (*ClientWatcher, error) {
	startReply := make(chan genWatchReply, 1)
	defer close(startReply)
	out := make(chan ResourceChanged, 64)

	select {
	case c.dataPlane <- &genWatcher{
		feedContext: ctx,
		feedChannel: out,
		started:     startReply,
	}:
	case <-ctx.Done():
		close(out)
		return nil, ctx.Err()
	}

	select {
	case reply := <-startReply:
		if reply.problem != nil {
			return nil, reply.problem
		}
		return &ClientWatcher{
			nextID:     0,
			res:        c,
			watcherID:  reply.id,
			Feed:       out,
			dispatches: make(map[WatchToken]watcherDispatch),
		}, nil
	case <-ctx.Done():
		close(out)
		return nil, ctx.Err()
	}
}

func (c *Client) List(ctx context.Context, kind Type) ([]Meta, error) {
	resultSignal := make(chan Meta, 16)
	select {
	case c.dataPlane <- &listOp{
		kind:    kind,
		replyTo: resultSignal,
	}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	var out []Meta
	for {
		select {
		case i, ok := <-resultSignal:
			if !ok {
				return out, nil
			}
			out = append(out, i)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
