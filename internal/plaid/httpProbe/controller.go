package httpProbe

import (
	"context"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources/operator"
	"github.com/thejerf/suture/v4"
	"time"
)

type Controller struct {
	resources *resources.Controller
	probeTree *suture.Supervisor
}

func NewController(res *resources.Controller, probeTree *suture.Supervisor) *Controller {
	return &Controller{
		resources: res,
		probeTree: probeTree,
	}
}

func (c *Controller) Serve(ctx context.Context) error {
	client := c.resources.Client()
	watcher, err := client.Watcher(ctx)
	if err != nil {
		return err
	}

	probeEvents, probeScheduler := newScheduler(c.probeTree)

	a1 := operator.NewKindBridge[AlphaV1Spec, AlphaV1Status, alphaV1Probe](AlphaV1Type, &alphaV1Interpreter{
		resources: client,
		watcher:   watcher,
		scheduler: probeScheduler,
	})
	a1Events, err := a1.Setup(ctx, client)
	if err != nil {
		return err
	}

	for {
		select {
		case e := <-watcher.Feed:
			if err := watcher.Digest(ctx, e); err != nil {
				return err
			}
		case e := <-a1Events:
			if err := a1.Dispatch(ctx, client, e); err != nil {
				return err
			}
		case e := <-probeEvents:
			if err := probeScheduler.consume(ctx, e); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type alphaV1Probe struct {
	self resources.Meta
	//todo: need to be able to unschedule
	scheduled bool
}

type alphaV1Interpreter struct {
	resources *resources.Client
	watcher   *resources.ClientWatcher
	scheduler *scheduler
}

func (a *alphaV1Interpreter) reconcile(ctx context.Context, p *alphaV1Probe, spec AlphaV1Spec) (AlphaV1Status, error) {
	status := AlphaV1Status{Ready: false}

	if !spec.Enabled {
		//a.logging <- fmt.Sprintf("[http-probe]\t%s:\tNot enabled.", p.self.Name)
		return status, nil
	}

	if !p.scheduled {
		a.scheduler.schedule(250*time.Millisecond, p.self, spec.Host, spec.Port, spec.Resource, func(ctx context.Context, result probeResult) error {
			return a.probeUpdate(ctx, p, result)
		})
		p.scheduled = true
	}
	return status, nil
}

func (a *alphaV1Interpreter) probeUpdate(ctx context.Context, p *alphaV1Probe, result probeResult) error {
	if _, err := a.resources.UpdateStatus(ctx, p.self, AlphaV1Status{Ready: result.success}); err != nil {
		return err
	} else {
		return nil
	}
}

func (a *alphaV1Interpreter) reconcileStatus(ctx context.Context, p *alphaV1Probe, spec AlphaV1Spec) error {
	status, err := a.reconcile(ctx, p, spec)
	if exists, err := a.resources.UpdateStatus(ctx, p.self, status); err != nil {
		return err //double error, not sure what to do
	} else if !exists {
		//deleted?
	}
	return err
}

func (a *alphaV1Interpreter) Create(ctx context.Context, which resources.Meta, spec AlphaV1Spec, bridgeState *operator.KindBridgeState) (*alphaV1Probe, AlphaV1Status, error) {
	//a.logging <- fmt.Sprintf("[http-probe]\t%s:\tCreating resource for %#v", which.Name, spec)
	p := &alphaV1Probe{
		self: which,
	}
	stat, err := a.reconcile(ctx, p, spec)
	return p, stat, err
}

func (a *alphaV1Interpreter) Update(ctx context.Context, which resources.Meta, rt *alphaV1Probe, s AlphaV1Spec) (AlphaV1Status, error) {
	return a.reconcile(ctx, rt, s)
}
