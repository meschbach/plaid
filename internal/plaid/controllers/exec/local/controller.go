package local

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/internal/plaid/controllers/logdrain"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/operator"
	"github.com/thejerf/suture/v4"
)

type Controller struct {
	storage         *resources.Controller
	procSupervisors *suture.Supervisor
	logging         *logdrain.ServiceConfig
}

func (c *Controller) Serve(ctx context.Context) error {
	store := c.storage.Client()

	bridge := operator.NewKindBridge[exec.InvocationAlphaV1Spec, exec.InvocationAlphaV1Status, proc](exec.InvocationAlphaV1Type, &alphaV1Ops{
		supervisor: c.procSupervisors,
		logging:    c.logging,
	})
	av1Event, err := bridge.Setup(ctx, store)
	if err != nil {
		return err
	}

	for {
		select {
		case e := <-av1Event:
			if err := bridge.Dispatch(ctx, store, e); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
