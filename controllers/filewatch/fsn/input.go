package fsn

import "context"

type inputOpCode uint8

const (
	addWatch inputOpCode = iota
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
	case <-ctx.Done():
	}

	return nil
}

func (c *Core) Watch(ctx context.Context, path string) error {
	return c.Watch2(ctx, WatchConfig{
		Path:          path,
		ExcludeSuffix: nil,
	})
}
