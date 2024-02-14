package project

import (
	"context"
	"errors"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/operator"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
	ctx, span := tracer.Start(parent, "project.Update", trace.WithAttributes(
		attribute.Stringer("name", which),
		attribute.Int("spec.oneshots", len(spec.OneShots)),
		attribute.Int("spec.daemons", len(spec.Daemons)),
	))
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
	span.SetAttributes(attribute.Int("one-shots.incomplete", incompleteOneShots))
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
			span.SetStatus(codes.Error, "failure while determine daemon state")
			span.RecordError(err)
			daemonErrors = append(daemonErrors, err)
		} else {
			subController.toStatus(daemonSpec, daemonStatus)
			switch next {
			case daemonWait:
				span.AddEvent("daemon-wait", trace.WithAttributes(attribute.Bool("daemon.ready", daemonStatus.Ready)))
				if !subController.targetReady {
					allDaemonsReady = false
				}
				//do nothing
			case daemonCreate:
				span.AddEvent("creating-daemon")
				err := subController.create(ctx, env, spec, daemonSpec)
				if err != nil {
					span.SetStatus(codes.Error, "failed to create daemon")
					span.RecordError(err)
					daemonErrors = append(daemonErrors, err)
					continue
				}
				allDaemonsReady = false
			case daemonFinished:
				span.AddEvent("daemon-finished")
				//todo: restart?
			}
			status.Daemons = append(status.Daemons, daemonStatus)
		}
	}
	span.SetAttributes(attribute.Bool("daemons.ready", allDaemonsReady))

	//todo: clean up tests and put this under test
	status.Ready = incompleteOneShots == 0 && allDaemonsReady
	allErrors := append(oneShotErrors, daemonErrors...)
	return status, errors.Join(allErrors...)
}

func (a *alpha1Ops) Delete(ctx context.Context, which resources.Meta, rt *state) error {
	env := &resourceEnv{
		which:   which,
		rpc:     a.client,
		watcher: a.watcher,
	}

	var problems []error
	for _, oneShot := range rt.oneShots {
		problems = append(problems, oneShot.delete(ctx, env))
	}
	for _, daemon := range rt.daemons {
		problems = append(problems, daemon.delete(ctx, env))
	}
	return errors.Join(problems...)
}

type state struct {
	bridge   *operator.KindBridgeState
	oneShots map[string]*oneShotState
	daemons  map[string]*daemonState
}
