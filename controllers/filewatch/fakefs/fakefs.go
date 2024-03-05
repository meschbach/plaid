// Package fakefs provides a virtual implementation of the filewatch system to aid in the development of system.
package fakefs

import (
	"context"
	"fmt"
	"github.com/meschbach/go-junk-bucket/pkg/fx"
	"github.com/meschbach/plaid/controllers/filewatch"
	"go.opentelemetry.io/otel/trace"
	"time"
)

type Core struct {
	ops            chan func(ctx context.Context, core *Core) error
	watchingPrefix []string
	output         chan filewatch.ChangeEvent
}

func (c *Core) Watch(ctx context.Context, path string) error {
	op := func(ctx context.Context, core *Core) error {
		core.watchingPrefix = append(core.watchingPrefix, path)
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.ops <- op:
		return nil
	}
}

func (c *Core) Unwatch(ctx context.Context, path string) error {
	op := func(ctx context.Context, core *Core) error {
		core.watchingPrefix = fx.Filter(core.watchingPrefix, func(e string) bool {
			return e != path
		})
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.ops <- op:
		return nil
	}
}

func (c *Core) ChangeFeed() <-chan filewatch.ChangeEvent {
	return c.output
}

func (c *Core) Serve(ctx context.Context) error {
	for {
		select {
		case op := <-c.ops:
			if err := op(ctx, c); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *Core) FileModified(ctx context.Context, filePath string, when time.Time) error {
	s := trace.SpanContextFromContext(ctx)
	op := func(ctx context.Context, core *Core) error {
		core.output <- filewatch.ChangeEvent{
			Kind:    filewatch.FileModified,
			Path:    filePath,
			When:    when,
			Tracing: s,
		}
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.ops <- op:
		return nil
	}
}

func New() *Core {
	return &Core{
		ops:    make(chan func(ctx context.Context, core *Core) error, 16),
		output: make(chan filewatch.ChangeEvent, 16),
	}
}
