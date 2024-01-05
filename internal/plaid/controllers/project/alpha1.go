package project

import (
	"context"
	"errors"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources/operator"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const annotationRunType = Kind + ":type"
const annotationRunTypeOneShot = "one-shot"
const annotationRunTypeDaemon = "daemon"
const annotationName = Kind + ":name"

type alpha1Ops struct {
	client  *resources.Client
	watcher *resources.ClientWatcher
}

func (a *alpha1Ops) Create(ctx context.Context, which resources.Meta, spec Alpha1Spec, bridge *operator.KindBridgeState) (*state, Alpha1Status, error) {
	runtime := &state{
		bridge:   bridge,
		oneShots: make(map[string]*oneShotState),
		daemons:  make(map[string]*daemonState),
	}
	status, err := a.Update(ctx, which, runtime, spec)
	return runtime, status, err
}

func (a *alpha1Ops) Update(parent context.Context, which resources.Meta, rt *state, spec Alpha1Spec) (Alpha1Status, error) {
	ctx, span := tracer.Start(parent, "project.Update", trace.WithAttributes(attribute.String("name", which.Name)))
	defer span.End()

	status := Alpha1Status{}

	//for each one shot
	env := &resourceEnv{
		which:   which,
		rpc:     a.client,
		watcher: a.watcher,
		reconcile: func(ctx context.Context) error {
			return rt.bridge.OnResourceChange(ctx, which)
		},
	}

	var oneShotErrors []error
	incompleteOneShots := 0
	failedOneShots := 0
	for _, oneShotSpec := range spec.OneShots {
		oneShotStatus := Alpha1OneShotStatus{
			Name: oneShotSpec.Name,
			Done: false,
		}
		var subController *oneShotState
		if osc, has := rt.oneShots[oneShotSpec.Name]; has {
			subController = osc
		} else {
			subController = &oneShotState{}
			rt.oneShots[oneShotSpec.Name] = subController
		}
		if next, err := subController.decideNextStep(ctx, env); err != nil {
			oneShotErrors = append(oneShotErrors, err)
		} else {
			subController.toStatus(&oneShotStatus)
			switch next {
			case oneShotWait:
				incompleteOneShots++
			case oneShotCreate:
				incompleteOneShots++
				err := subController.create(ctx, env, spec, oneShotSpec)
				if err != nil {
					oneShotErrors = append(oneShotErrors, err)
				}
				subController.toStatus(&oneShotStatus)
			case oneShotFinished:
				if subController.finishState == oneShotFailure {
					failedOneShots++
				}
				oneShotStatus.Done = true
			}
		}
		status.OneShots = append(status.OneShots, oneShotStatus)
	}
	status.Done = incompleteOneShots == 0 && len(spec.Daemons) == 0
	if status.Done {
		if failedOneShots > 0 {
			status.Result = Alpha1StateFailed
		} else {
			status.Result = Alpha1StateSuccess
		}
	}

	//for each daemon
	allDaemonsReady := true
	var daemonErrors []error
	for _, daemonSpec := range spec.Daemons {
		var subController *daemonState
		if osc, has := rt.daemons[daemonSpec.Name]; has {
			subController = osc
		} else {
			subController = &daemonState{}
			rt.daemons[daemonSpec.Name] = subController
		}

		daemonStatus := &Alpha1DaemonStatus{}
		if next, err := subController.decideNextStep(ctx, env); err != nil {
			daemonErrors = append(daemonErrors, err)
		} else {
			switch next {
			case daemonWait:
				//do nothing
			case daemonCreate:
				err := subController.create(ctx, env, spec, daemonSpec)
				if err != nil {
					daemonErrors = append(daemonErrors, err)
				}
			case daemonFinished:
				//todo: restart?
			}
			subController.toStatus(daemonSpec, daemonStatus)
			allDaemonsReady = allDaemonsReady && daemonStatus.Ready
			status.Daemons = append(status.Daemons, daemonStatus)
		}
	}

	//todo: clean up tests and put this under test
	status.Ready = incompleteOneShots == 0 && allDaemonsReady
	allErrors := append(oneShotErrors, daemonErrors...)
	return status, errors.Join(allErrors...)
}

type state struct {
	bridge *operator.KindBridgeState
	//todo: watches used anywhere?
	watches  map[resources.Meta]resources.WatchToken
	oneShots map[string]*oneShotState
	daemons  map[string]*daemonState
}
