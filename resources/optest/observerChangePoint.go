package optest

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

type ObserverChangePoint struct {
	aspect *ObserverAspect
	origin eventCounter
}

func (r *ObserverChangePoint) Wait(t *testing.T, ctx context.Context, failedMessageAndArgs ...any) {
	t.Helper()
	for r.origin >= r.aspect.events() {
		err := r.aspect.consumeEvent(ctx)
		if err != nil {
			require.NoError(t, err, failedMessageAndArgs...)
			return
		}
	}
}

func (r *ObserverChangePoint) WaitFor(t *testing.T, ctx context.Context, satisfied func(ctx context.Context) bool, failedMessageAndArgs ...any) {
	for !satisfied(ctx) {
		r.Wait(t, ctx, failedMessageAndArgs...)
	}
}
