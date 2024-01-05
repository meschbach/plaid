package resources

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type dataOp interface {
	perform(ctx context.Context, c *Controller)
}

type Controller struct {
	id        int
	dataPlane chan dataOp
	//1.19.2
	//BenchmarkControllerGetOrCreateNode-8     2458381               576.7 ns/op           425 B/op          4 allocs/op
	//BenchmarkControllerGetOrCreateNode-8     2432014               501.3 ns/op           426 B/op          4 allocs/op
	//BenchmarkControllerGetOrCreateNode-8     2651074               522.8 ns/op           418 B/op          4 allocs/op
	//1.20.2
	//BenchmarkControllerGetOrCreateNode-8     2624342               449.4 ns/op           419 B/op          4 allocs/op
	//BenchmarkControllerGetOrCreateNode-8     2825480               463.3 ns/op           413 B/op          4 allocs/op
	//BenchmarkControllerGetOrCreateNode-8     2412211               473.9 ns/op           427 B/op          4 allocs/op
	resources   MetaContainer[node]
	allWatchers []*watcher
}

func (r *Controller) Serve(ctx context.Context) error {
	for {
		select {
		case op := <-r.dataPlane:
			op.perform(ctx, r)
		case <-ctx.Done():
			return nil
		}
	}
}

func (r *Controller) Client() *Client {
	return &Client{
		dataPlane: r.dataPlane,
	}
}

func (r *Controller) createOrGetNode(ctx context.Context, meta Meta) (*node, bool, error) {
	node, created := r.resources.GetOrCreate(meta, func() *node {
		return &node{}
	})
	return node, created, nil
}

// node represents the internal state of resources for the system.
type node struct {
	// spec is the desired state of the target resources.  Operators should consume the spec and attempt to
	// rectify the state.
	spec []byte
	// status is the current state of the resource from an operators perspective.  This means the status by necessity
	// lags behind spec, being eventually consistent.
	status []byte
	// events are notes made on resources made by the operator over time.  Might be cases for other resources to
	// append however this should probably be rare.
	events []Event
	//claimedBy is a slice of objects 'owning' this object, responsible for the lifecycle of the object
	claimedBy []Meta
	//annotations are additional meta data associated with the element
	annotations map[string]string
}

// getNode locates the given meta resource within the entity states.  If the referenced resource does not exist then
// nil will be returned.
func (r *Controller) getNode(ctx context.Context, meta Meta) (*node, error) {
	node, has := r.resources.Find(meta)
	if !has {
		return nil, nil
	}
	return node, nil
}

func (r *Controller) deleteNode(ctx context.Context, what Meta) (bool, error) {
	_, has := r.resources.Delete(what)
	return has, nil
}

func (r *Controller) dispatchUpdates(parent context.Context, changed ResourceChangedOperation, what Meta) {
	ctx, span := tracing.Start(parent, "Controller.dispatchUpdates",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(attribute.Stringer("change.operation", changed), attribute.Stringer("change.which", what)))
	defer span.End()

	link := trace.LinkFromContext(ctx)

	for _, w := range r.allWatchers {
		//todo: cull stale watchers
		w.dispatch(ctx, changed, what, link)
	}
}

func NewController() *Controller {
	return &Controller{
		id:        0,
		dataPlane: make(chan dataOp, 128),
	}
}
