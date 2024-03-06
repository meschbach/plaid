package filewatch

import (
	"context"
	"github.com/meschbach/plaid/controllers/tooling/kit"
	"github.com/meschbach/plaid/resources"
)

type Controller struct {
	resources resources.System
	watcher   FileSystem
}

func (c *Controller) Serve(ctx context.Context) error {
	storage, err := c.resources.Storage(ctx)
	if err != nil {
		return err
	}
	storageWatcher, err := storage.Observer(ctx)
	if err != nil {
		return err
	}
	storageFeed := storageWatcher.Events()

	rt := &runtimeState{
		resources: storage,
		fs:        c.watcher,
	}

	bridge := kit.New[Alpha1Spec, Alpha1Status, watch](storage, storageWatcher, Alpha1, &alpha1Interpreter{runtime: rt})
	if err := bridge.Setup(ctx); err != nil {
		return err
	}
	fsChangeFeed := c.watcher.ChangeFeed()

	for {
		select {
		case event := <-fsChangeFeed:
			if err := rt.consumeFSEvent(ctx, event); err != nil {
				return err
			}
		case v1Change := <-storageFeed:
			if err := storageWatcher.Digest(ctx, v1Change); err != nil {
				return err
			}
		case e := <-bridge.Loopback:
			if err := bridge.DigestLoopback(ctx, e); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func NewController(r resources.System, watcher FileSystem) *Controller {
	return &Controller{resources: r, watcher: watcher}
}
