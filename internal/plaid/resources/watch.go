package resources

import (
	"context"
	"go.opentelemetry.io/otel/trace"
)

type watchStartedReply struct {
	problem error
}

type watchStartOp struct {
	feedContext context.Context
	feedChannel chan ResourceChanged
	what        Type
	started     chan<- watchStartedReply
}

func (w *watchStartOp) perform(ctx context.Context, c *Controller) {
	c.allWatchers = append(c.allWatchers, &watcher{
		ctx:     w.feedContext,
		feed:    w.feedChannel,
		filters: []watchFilter{&typeWatchFilter{filterFor: w.what}},
		done:    false,
	})
	select {
	case <-ctx.Done():
		return
	case w.started <- watchStartedReply{problem: nil}:
	}
}

// watcher is the resource service side of a watch
type watcher struct {
	id   int
	ctx  context.Context
	feed chan ResourceChanged

	filters []watchFilter

	done bool
}

func (w *watcher) dispatch(ctx context.Context, op ResourceChangedOperation, meta Meta, dispatchLink trace.Link) {
	if w.done {
		return
	}
	matches := false
	for _, f := range w.filters {
		matches = f.matches(op, meta)
		if matches {
			break
		}
	}
	if !matches {
		return
	}

	select {
	case <-ctx.Done():
		return
	case <-w.ctx.Done():
		w.done = true
		return
	case w.feed <- ResourceChanged{
		Which:     meta,
		Operation: op,
		Tracing:   dispatchLink,
	}:
		return
	default:
		//todo: figure out how to handle full client queues
	}
}

type watchFilter interface {
	matches(op ResourceChangedOperation, which Meta) bool
}

type typeWatchFilter struct {
	filterFor Type
}

func (t *typeWatchFilter) matches(op ResourceChangedOperation, which Meta) bool {
	if t.filterFor.Kind != which.Type.Kind {
		return false
	}
	if t.filterFor.Version != which.Type.Version {
		return false
	}
	return true
}

type exactMatch struct {
	target Meta
}

func (e *exactMatch) matches(op ResourceChangedOperation, which Meta) bool {
	return e.target == which
}
