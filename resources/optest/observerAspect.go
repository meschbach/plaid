package optest

import "context"

type ObserverAspect struct {
	observer  *Observer
	seenCount eventCounter
}

func (a *ObserverAspect) events() eventCounter {
	return a.seenCount
}

func (a *ObserverAspect) consumeEvent(ctx context.Context) error {
	return a.observer.consumeEvent(ctx)
}

func (a *ObserverAspect) update() {
	a.seenCount++
}

func (a *ObserverAspect) Fork() *ObserverChangePoint {
	return &ObserverChangePoint{
		aspect: a,
		origin: a.seenCount,
	}
}
