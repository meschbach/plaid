package service

import (
	"context"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/operator"
	"time"
)

type Controller struct {
	storage *resources.Controller
}

func NewSystem(storage *resources.Controller) *Controller {
	return &Controller{storage: storage}
}

func (c *Controller) Serve(ctx context.Context) error {
	store := c.storage.Client()

	watcher, err := store.Watcher(ctx)
	if err != nil {
		return err
	}

	bridge := operator.NewKindBridge[Alpha1Spec, Alpha1Status, serviceState](Alpha1, &alpha1Ops{
		client:  store,
		watcher: watcher,
	})
	av1Event, err := bridge.Setup(ctx, store)
	if err != nil {
		return err
	}

	rescanTimer := time.NewTicker(30 * time.Second)
	defer rescanTimer.Stop()

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
		case <-rescanTimer.C:
			if err := bridge.Rescan(ctx); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
