package fsn

import "context"

type inputOpCode uint8

const (
	addWatch inputOpCode = iota
	removeWatch
)

type inputOp struct {
	watch WatchConfig
	op    inputOpCode
}

// WatchConfig specifies the configuration for a specific watch
type WatchConfig struct {
	// Path is the root of the watch
	Path string
	// ExcludeSuffix will ignore directories matching this suffix.  For example `.git`, `.idea`, and `~`.
	ExcludeSuffix []string
}

// todo: Update FileSystem to use WatchPoint
func (c *Core) Watch2(ctx context.Context, point WatchConfig) error {
	select {
	case c.input <- inputOp{
		watch: point,
		op:    addWatch,
	}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Core) Watch(ctx context.Context, path string) error {
	return c.Watch2(ctx, WatchConfig{
		Path:          path,
		ExcludeSuffix: []string{"node_modules", ".git"},
	})
}

func (c *Core) Unwatch(ctx context.Context, path string) error {
	select {
	case c.input <- inputOp{
		watch: WatchConfig{
			Path: path,
		},
		op: removeWatch,
	}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
