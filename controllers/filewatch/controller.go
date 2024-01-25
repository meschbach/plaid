package filewatch

import (
	"context"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/operator"
)

type Controller struct {
	resources *resources.Controller
	watcher   FileSystem
}

func (c *Controller) Serve(ctx context.Context) error {
	rt := &runtimeState{
		fs: c.watcher,
	}

	resourceClient := c.resources.Client()
	rt.resources = resourceClient

	a1Bridge := operator.NewKindBridge[Alpha1Spec, Alpha1Status, watch](Alpha1, &alpha1Interpreter{runtime: rt})
	a1Watch, err := a1Bridge.Setup(ctx, resourceClient)
	if err != nil {
		return err
	}
	fsChangeFeed := c.watcher.ChangeFeed()

	for {
		select {
		case event := <-fsChangeFeed:
			if err := rt.consumeFSEvent(ctx, event); err != nil {
				return err
			}
		case v1Change := <-a1Watch:
			if err := a1Bridge.Dispatch(ctx, resourceClient, v1Change); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func NewController(r *resources.Controller, watcher FileSystem) *Controller {
	return &Controller{resources: r, watcher: watcher}
}
