package project

import (
	"context"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/operator"
)

type Controller struct {
	storage *resources.Controller
}

func NewProjectSystem(storage *resources.Controller) *Controller {
	return &Controller{storage: storage}
}

func (c *Controller) Serve(ctx context.Context) error {
	store := c.storage.Client()

	watcher, err := store.Watcher(ctx)
	if err != nil {
		return err
	}

	bridge := operator.NewKindBridge[Alpha1Spec, Alpha1Status, state](Alpha1, &alpha1Ops{
		client:  store,
		watcher: watcher,
	})
	av1Event, err := bridge.Setup(ctx, store)
	if err != nil {
		return err
	}

	for {
		select {
		case e := <-watcher.Feed:
			if err := watcher.Digest(ctx, e); err != nil {
				return err
			}
		case e := <-av1Event:
			if err := bridge.Dispatch(ctx, store, e); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
