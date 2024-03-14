package project

import (
	"context"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/operator"
	"time"
)

type ControllerOpts struct {
	//DefaultWatchFiles provides the default value when a project is missing the WatchFiles field is missing
	DefaultWatchFiles bool
}

type Controller struct {
	storage           *resources.Controller
	defaultWatchFiles bool
}

func NewProjectSystem(storage *resources.Controller, opts ControllerOpts) *Controller {
	return &Controller{
		storage:           storage,
		defaultWatchFiles: opts.DefaultWatchFiles,
	}
}

func (c *Controller) Serve(ctx context.Context) error {
	store := c.storage.Client()

	watcher, err := store.Watcher(ctx)
	if err != nil {
		return err
	}

	bridge := operator.NewKindBridge[Alpha1Spec, Alpha1Status, state](Alpha1, &alpha1Ops{
		client:            store,
		watcher:           watcher,
		defaultWatchFiles: c.defaultWatchFiles,
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
