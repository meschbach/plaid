package projectfile

import (
	"context"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources/operator"
)

const Kind = "plaid.meschbach.com/project-file"

type Controller struct {
	storage *resources.Controller
}

func (c *Controller) Serve(controllerContext context.Context) error {
	store := c.storage.Client()

	watcher, err := store.Watcher(controllerContext)
	if err != nil {
		return err
	}

	bridge := operator.NewKindBridge[Alpha1Spec, Alpha1Status, state](Alpha1, &alpha1Ops{
		storage: store,
		watcher: watcher,
	})
	av1Event, err := bridge.Setup(controllerContext, store)
	if err != nil {
		return err
	}

	for {
		select {
		case w := <-watcher.Feed:
			if err := watcher.Digest(controllerContext, w); err != nil {
				return err
			}
		case e := <-av1Event:
			if err := bridge.Dispatch(controllerContext, store, e); err != nil {
				return err
			}
		case <-controllerContext.Done():
			return controllerContext.Err()
		}
	}
}

func NewProjectFileSystem(storage *resources.Controller) *Controller {
	return &Controller{storage: storage}
}
