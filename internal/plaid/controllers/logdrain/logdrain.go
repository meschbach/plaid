package logdrain

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/reactors"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
	"github.com/meschbach/plaid/resources"
	"github.com/thejerf/suture/v4"
)

type ServiceConfig struct {
	core         *core
	coreBoundary reactors.Boundary[*core]
}

func BuildClient[T any](ctx context.Context, builder *ServiceConfig, from reactors.Boundary[T]) *Client[T] {
	return &Client[T]{
		clientWell:   from,
		logDrainWell: builder.coreBoundary,
	}
}

func NewLogDrainSystem(res *resources.Controller) (*ServiceConfig, *suture.Supervisor) {
	coreReactor, coreReactorInput := reactors.NewChannel[*core](10)

	core := &core{
		output: streams.NewFanOutSink[LogEntry](),
	}

	controller := &Controller{
		storage:           res,
		core:              core,
		coreReactor:       coreReactor,
		coreReactorEvents: coreReactorInput,
	}

	logging := suture.NewSimple("log-drains")
	logging.Add(controller)

	config := &ServiceConfig{
		core:         core,
		coreBoundary: coreReactor,
	}
	return config, logging
}
