// Package fsn wraps fsnotify package to integrate with operating system watching
package fsn

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/meschbach/plaid/controllers/filewatch"
	"io/fs"
	"path/filepath"
	"strings"
)

type Core struct {
	input  chan inputOp
	Output chan filewatch.ChangeEvent
}

func NewCore() *Core {
	return &Core{
		input:  make(chan inputOp, 4),
		Output: make(chan filewatch.ChangeEvent, 16),
	}
}

func (c *Core) Serve(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	for {
		select {
		case op := <-c.input:
			switch op.op {
			case addWatch:
				//todo: notify sub-watches
				if err := filepath.WalkDir(op.watch.Path, func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if d.IsDir() {
						name := d.Name()
						for _, suffix := range op.watch.ExcludeSuffix {
							if strings.HasSuffix(name, suffix) {
								return filepath.SkipDir
							}
						}
						if err := watcher.AddWith(path); err != nil {
							return err
						}
					}
					return nil
				}); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unknown operation %d", op.op)
			}
		case err := <-watcher.Errors:
			return err
		case e := <-watcher.Events:
			if err := c.consumeFSEvent(ctx, e); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
