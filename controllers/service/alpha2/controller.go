package alpha2

import (
	"context"
	"github.com/meschbach/plaid/controllers/tooling/kit"
	"github.com/meschbach/plaid/resources"
)

func NewController(system resources.System) *Controller {
	return &Controller{system: system}
}

type Controller struct {
	system resources.System
}

func (c *Controller) Serve(ctx context.Context) error {
	storage, err := c.system.Storage(ctx)
	if err != nil {
		return err
	}
	watcher, err := storage.Observer(ctx)
	if err != nil {
		return err
	}

	alpha2Kit := kit.New[Spec, Status, State](storage, watcher, Type, &Ops{
		storage:  storage,
		observer: watcher,
	})
	if err := alpha2Kit.Setup(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-watcher.Events():
			if err := watcher.Digest(ctx, e); err != nil {
				return err
			}
		case a := <-alpha2Kit.Loopback:
			if err := alpha2Kit.DigestLoopback(ctx, a); err != nil {
				return err
			}
		}
	}
}
