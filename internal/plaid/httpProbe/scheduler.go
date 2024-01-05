package httpProbe

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/resources"
	"github.com/thejerf/suture/v4"
	"time"
)

type scheduler struct {
	subtree *suture.Supervisor
	//todo: a lot of these should probably just go into a single container of id -> {obj}
	probes   map[resources.Meta]*probe
	consumer map[resources.Meta]func(ctx context.Context, result probeResult) error
	tokens   map[resources.Meta]suture.ServiceToken
	results  chan probeResult
}

func newScheduler(subtree *suture.Supervisor) (chan probeResult, *scheduler) {
	channel := make(chan probeResult, 32)
	return channel, &scheduler{
		subtree:  subtree,
		probes:   make(map[resources.Meta]*probe),
		consumer: make(map[resources.Meta]func(context.Context, probeResult) error),
		tokens:   make(map[resources.Meta]suture.ServiceToken),
		results:  channel,
	}
}

func (s *scheduler) consume(ctx context.Context, result probeResult) error {
	if consumer, ok := s.consumer[result.id]; ok {
		return consumer(ctx, result)
	} else { //last message from completed?
		return nil
	}
}

func (s *scheduler) schedule(period time.Duration, id resources.Meta, host string, port uint16, resource string, onResult func(ctx context.Context, result probeResult) error) {
	timer := time.NewTicker(period)
	p := &probe{
		id:       id,
		period:   timer,
		host:     host,
		port:     port,
		resource: resource,
		notify:   s.results,
	}
	//todo: handle case one already exists
	s.probes[id] = p
	s.consumer[id] = onResult
	s.tokens[id] = s.subtree.Add(p)
}
