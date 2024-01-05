package registry

import (
	"context"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/operator"
)

type Controller struct {
	resources *resources.Controller
}

func NewController(r *resources.Controller) *Controller {
	return &Controller{
		r,
	}
}

func (c *Controller) Serve(ctx context.Context) error {
	res := c.resources.Client()

	a1 := operator.NewKindBridge[AlphaV1Spec, AlphaV1Status, registry](AlphaV1, &alpha1Interpreter{res: res})
	a1Changes, err := a1.Setup(ctx, res)
	if err != nil {
		return err
	}

	for {
		select {
		case change := <-a1Changes:
			if err := a1.Dispatch(ctx, res, change); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
