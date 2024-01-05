package logger

import (
	"context"
	"errors"
	"github.com/meschbach/plaid/internal/plaid/controllers/logdrain"
	"github.com/meschbach/plaid/internal/plaid/resources"
	"github.com/meschbach/go-junk-bucket/pkg/reactors"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
	"github.com/meschbach/go-junk-bucket/pkg/task"
	"sync"
)

type ExportedLogDrain struct {
	storage *resources.Controller
	config  *logdrain.ServiceConfig

	reactor      *reactors.Channel[*exportedDrainState]
	reactorQueue <-chan reactors.ChannelEvent[*exportedDrainState]
}

func (e *ExportedLogDrain) registerDrain(parent context.Context, name string, output streams.Sink[bufferedEntry]) (bool, error) {
	ctx, span := tracing.Start(parent, "ExportedLogDrain.registerDrain")
	defer span.End()

	gate := sync.WaitGroup{}
	gate.Add(1)
	var problem error
	e.reactor.ScheduleStateFunc(ctx, func(ctx context.Context, state *exportedDrainState) error {
		adapter := streams.WrapTransformingSink[logdrain.LogEntry, bufferedEntry](output, func(ctx context.Context, in logdrain.LogEntry) (bufferedEntry, error) {
			return bufferedEntry{
				when:       in.When,
				text:       in.Message,
				from:       in.Origin.From,
				streamName: in.Origin.Stream,
			}, nil
		})
		//todo: hardwired drain name
		p := state.logging.RegisterDrain(ctx, adapter)
		p.OnCompleted(ctx, func(ctx context.Context, event task.Result[bool]) {
			defer gate.Done()
			if event.Problem != nil {
				problem = event.Problem
			}
		})
		return nil
	})
	gate.Wait()

	return problem == nil, problem
}

func (e *ExportedLogDrain) unregisterDrain(ctx context.Context, name string) error {
	return errors.New("TODO")
}

type exportedDrainState struct {
	logging *logdrain.Client[*exportedDrainState]
}

func (e *ExportedLogDrain) Serve(ctx context.Context) error {
	logging := logdrain.BuildClient[*exportedDrainState](ctx, e.config, e.reactor)
	state := &exportedDrainState{
		logging: logging,
	}

	for {
		select {
		case event := <-e.reactorQueue:
			if err := e.reactor.Tick(ctx, event, state); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func ExportV1LogDrain(storage *resources.Controller, drainConfig *logdrain.ServiceConfig) (*ExportedLogDrain, V1Server) {
	reactor, input := reactors.NewChannel[*exportedDrainState](16)
	export := &ExportedLogDrain{
		reactor:      reactor,
		reactorQueue: input,
		storage:      storage,
		config:       drainConfig,
	}
	service := newV1LoggingService(export)
	return export, service
}
