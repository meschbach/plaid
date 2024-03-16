package resources

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/fx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type WatchToken uint

// ClientWatcher provides a client interface to a filterable change feed
type ClientWatcher struct {
	res *Client
	//watcherID is the ID provided by the Client
	watcherID  int
	Feed       chan ResourceChanged
	dispatches map[WatchToken]watcherDispatch
	nextID     WatchToken
}

// OnResourceChanged is a handler for a resource event
type OnResourceChanged func(ctx context.Context, changed ResourceChanged) error

type watcherDispatch struct {
	filter watchFilter
	target OnResourceChanged
}

func (w *watcherDispatch) dispatch(ctx context.Context, event ResourceChanged) error {
	return w.target(ctx, event)
}

func (c *ClientWatcher) Digest(parent context.Context, event ResourceChanged) error {
	ctx, span := TracedMessageConsumer(parent, event.Tracing, "ClientWatcher.Digest: "+event.Operation.String()+" "+event.Which.Type.String(), trace.WithAttributes(event.ToTraceAttributes()...))
	defer span.End()

	total := len(c.dispatches)
	attempted := 0
	matched := 0
	defer func() {
		span.SetAttributes(
			attribute.Int("filters", total),
			attribute.Int("attempted", attempted),
			attribute.Int("matched", matched),
		)
	}()

	for _, d := range c.dispatches {
		attempted++
		if d.filter.matches(event.Operation, event.Which) {
			matched++

			if err := d.dispatch(ctx, event); err != nil {
				span.SetStatus(codes.Error, err.Error())
				return err
			}
		}
	}
	return nil
}

func (c *ClientWatcher) addFilter(filter watchFilter, consume OnResourceChanged) (WatchToken, error) {
	dispatch := watcherDispatch{
		filter: filter,
		target: consume,
	}
	id := c.nextID
	c.nextID++
	c.dispatches[id] = dispatch
	return id, nil
}

// WatchType watches for changes to kind and invokes consume when an event is triggered.
// Deprecated: Please use ClientWatcher.OnType to be able to turn off various watches
func (c *ClientWatcher) WatchType(ctx context.Context, kind Type, consume OnResourceChanged) error {
	_, err := c.OnType(ctx, kind, consume)
	return err
}

func (c *ClientWatcher) OnType(ctx context.Context, kind Type, consume OnResourceChanged) (WatchToken, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case c.res.dataPlane <- &watchAddTypeFilter{
		watchID: c.watcherID,
		forType: kind,
	}:
	}
	return c.addFilter(&typeWatchFilter{filterFor: kind}, consume)
}

func (c *ClientWatcher) WatchResource(ctx context.Context, ref Meta, consume OnResourceChanged) error {
	_, err := c.OnResource(ctx, ref, consume)
	return err
}

func (c *ClientWatcher) OnResource(parent context.Context, ref Meta, consume OnResourceChanged) (WatchToken, error) {
	ctx, span := tracing.Start(parent, "ClientWatcher.OnResource", trace.WithAttributes(ref.AsTraceAttribute("ref")...))
	defer span.End()
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case c.res.dataPlane <- &watchAddResourceFilter{
		watchID:     c.watcherID,
		forResource: ref,
	}:
	}
	return c.addFilter(&exactMatch{target: ref}, consume)
}

func (c *ClientWatcher) OnAll(ctx context.Context, consumer OnResourceChanged) (WatchToken, error) {
	token, err := c.addFilter(&watcherMatchAll{}, consumer)
	if err != nil {
		return 0, err
	}

	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case c.res.dataPlane <- &watcherAddAllOp{
		watchID: c.watcherID,
	}:
	}
	return token, err
}

func (c *ClientWatcher) Serve(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-c.Feed:
			if err := c.Digest(ctx, e); err != nil {
				return err
			}
		}
	}
}

type watcherAddStatusChangedOp struct {
	watchID int
	which   Meta
}

func (w *watcherAddStatusChangedOp) perform(ctx context.Context, c *Controller) {
	for _, watcher := range c.allWatchers {
		if watcher.id == w.watchID {
			watcher.filters = append(watcher.filters, &watcherMatchResourceOperation{
				which: w.which,
				op:    StatusUpdated,
			}, &watcherMatchResourceOperation{
				which: w.which,
				op:    DeletedEvent,
			})
		}
	}
}

// OnResourceStatusChanged registers consumer for invocation when either a status update occurs or a deletion operation
func (c *ClientWatcher) OnResourceStatusChanged(parent context.Context, which Meta, consumer OnResourceChanged) (WatchToken, error) {
	ctx, span := tracing.Start(parent, "ClientWatcher.OnResourceStatusChanged", trace.WithAttributes(which.AsTraceAttribute("which")...))
	defer span.End()

	filter := &watcherAnyOperation{
		to: []watchFilter{
			&watcherMatchResourceOperation{
				which: which,
				op:    StatusUpdated,
			},
			&watcherMatchResourceOperation{
				which: which,
				op:    DeletedEvent,
			},
		},
	}

	token, err := c.addFilter(filter, consumer)
	if err != nil {
		span.SetStatus(codes.Error, "failed to add filter")
		return 0, err
	}

	select {
	case <-ctx.Done():
		span.SetStatus(codes.Error, "context cancelled")
		return 0, ctx.Err()
	case c.res.dataPlane <- &watcherAddStatusChangedOp{
		watchID: c.watcherID,
		which:   which,
	}:
	}
	return token, nil
}

func (c *ClientWatcher) Off(ctx context.Context, token WatchToken) error {
	if token == 0 {
		return nil
	}
	delete(c.dispatches, token)
	return nil
}

func (c *ClientWatcher) Close(ctx context.Context) error {
	c.res.dataPlane <- &removeWatcher{id: c.watcherID}
	return nil
}

func (c *ClientWatcher) Events() chan ResourceChanged {
	return c.Feed
}

type genWatcher struct {
	feedContext context.Context
	feedChannel chan ResourceChanged
	started     chan<- genWatchReply
}

func (g *genWatcher) perform(ctx context.Context, c *Controller) {
	c.id++
	id := c.id
	c.allWatchers = append(c.allWatchers, &watcher{
		ctx:     g.feedContext,
		feed:    g.feedChannel,
		filters: nil,
		done:    false,
		id:      id,
	})
	select {
	case <-ctx.Done():
		return
	case g.started <- genWatchReply{id: id}:
	}
}

type genWatchReply struct {
	id      int
	problem error
}

// todo: optimize the watcher
type watcherMatchAll struct {
}

func (w *watcherMatchAll) matches(op ResourceChangedOperation, which Meta) bool {
	return true
}

type removeWatcher struct {
	id int
}

func (r *removeWatcher) perform(ctx context.Context, c *Controller) {
	matching, others := fx.Split(c.allWatchers, func(e *watcher) bool {
		return e.id == r.id
	})
	for _, m := range matching {
		m.done = true
	}
	c.allWatchers = others
}

type watchAddTypeFilter struct {
	watchID int
	forType Type
}

func (w *watchAddTypeFilter) perform(ctx context.Context, c *Controller) {
	//todo: might be faster to use a map?
	for _, watcher := range c.allWatchers {
		if watcher.id == w.watchID {
			watcher.filters = append(watcher.filters, &typeWatchFilter{filterFor: w.forType})
		}
	}
}

type watchAddResourceFilter struct {
	watchID     int
	forResource Meta
}

func (w *watchAddResourceFilter) perform(ctx context.Context, c *Controller) {
	for _, watcher := range c.allWatchers {
		if watcher.id == w.watchID {
			watcher.filters = append(watcher.filters, &exactMatch{target: w.forResource})
		}
	}
}

type watcherAddAllOp struct {
	watchID int
}

func (w *watcherAddAllOp) perform(ctx context.Context, c *Controller) {
	for _, watcher := range c.allWatchers {
		if watcher.id == w.watchID {
			watcher.filters = append(watcher.filters, &watcherMatchAll{})
		}
	}
}

type watcherMatchResourceOperation struct {
	which Meta
	op    ResourceChangedOperation
}

func (w *watcherMatchResourceOperation) matches(op ResourceChangedOperation, which Meta) bool {
	if !w.which.EqualsMeta(which) {
		return false
	}
	return w.op == op
}

type watcherAnyOperation struct {
	to []watchFilter
}

func (w *watcherAnyOperation) matches(op ResourceChangedOperation, which Meta) bool {
	for _, f := range w.to {
		if f.matches(op, which) {
			return true
		}
	}
	return false
}
