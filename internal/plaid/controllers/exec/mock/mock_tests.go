package mock

import (
	"context"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/controllers/logdrain"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"github.com/meschbach/go-junk-bucket/pkg/emitter"
	"github.com/meschbach/go-junk-bucket/pkg/reactors"
	"github.com/meschbach/go-junk-bucket/pkg/task"
)

func With[Outer any](ctx context.Context, replyTo reactors.Boundary[Outer], m *Proc, perform func(ctx context.Context, m *Proc) error) *task.Promise[bool] {
	return reactors.Submit[Outer, *engineState, bool](ctx, replyTo, m.interpreter.engine.reactor, func(boundaryContext context.Context, state *engineState) (bool, error) {
		return true, perform(boundaryContext, m)
	})
}

// AttachMockInvocationController creates a new execution engine capable of being controlled via testing.
func AttachMockInvocationController(ctx context.Context, core *resources.TestSubsystem, loggingConfig *logdrain.ServiceConfig) *ExecEngine {
	eventLoop, eventLoopInput := reactors.NewChannel[*engineState](128)

	//todo: probably wrong location to create logging client
	loggingClient := logdrain.BuildClient[*engineState](ctx, loggingConfig, eventLoop)
	mockInvocationController := &ExecEngine{
		c:         core.Controller,
		reactor:   eventLoop,
		reactorIn: eventLoopInput,
		events:    &emitter.Dispatcher[*engineState]{},
		logging:   loggingClient,
	}
	core.AttachController("mock-exec-engine", mockInvocationController)
	return mockInvocationController
}
