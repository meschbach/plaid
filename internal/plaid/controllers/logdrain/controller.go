package logdrain

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/resources"
	"github.com/meschbach/plaid/internal/plaid/resources/operator"
	"github.com/meschbach/go-junk-bucket/pkg/reactors"
)

type Controller struct {
	storage           *resources.Controller
	sourceRegistry    *Registry
	core              *core
	coreReactor       *reactors.Channel[*core]
	coreReactorEvents <-chan reactors.ChannelEvent[*core]
}

func (c *Controller) Serve(ctx context.Context) error {
	storage := c.storage.Client()
	a1Bridge := operator.NewKindBridge[Alpha1Spec, Alpha1Status, alpha1State](AlphaV1Type, &alphaV1{
		c: c,
	})
	av1Event, err := a1Bridge.Setup(ctx, storage)
	if err != nil {
		return err
	}

	for {
		select {
		case r := <-c.coreReactorEvents:
			if err := c.coreReactor.Tick(ctx, r, c.core); err != nil {
				return err
			}
		case e := <-av1Event:
			if err := a1Bridge.Dispatch(ctx, storage, e); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
