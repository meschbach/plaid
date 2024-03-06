package resources

import (
	"context"
	"time"
)

type EventLevel uint8

const (
	Error EventLevel = iota
	Info
	AllEvents
)

type Event struct {
	When   time.Time  `json:"when"`
	Level  EventLevel `json:"level"`
	Format string     `json:"format"`
	Params []any      `json:"params"`
}

type logEventReply struct {
	exists  bool
	problem error
}

type logEvent struct {
	replyTo chan logEventReply
	entity  Meta
	event   Event
}

func (l *logEvent) name() string {
	return "log-event"
}

func (l *logEvent) perform(ctx context.Context, c *Controller) {
	node, err := c.getNode(ctx, l.entity)
	if err != nil {
		l.replyTo <- logEventReply{problem: err}
		return
	}
	if node == nil {
		l.replyTo <- logEventReply{exists: false}
		return
	}
	node.events = append(node.events, l.event)
	l.replyTo <- logEventReply{exists: true}
	close(l.replyTo) //todo: does not allow reuse on repaid changes
}

// Log records an entry against the referenced entity.
func (c *Client) Log(ctx context.Context, entity Meta, level EventLevel, format string, args ...any) (bool, error) {
	if level == AllEvents {
		panic("level may not be all for logging")
	}
	result := make(chan logEventReply, 1)
	select {
	case c.dataPlane <- &logEvent{
		replyTo: result,
		entity:  entity,
		event: Event{
			When:   time.Now(),
			Level:  level,
			Format: format,
			Params: args,
		},
	}:
	case <-ctx.Done():
		return false, ctx.Err()
	}

	select {
	case r := <-result:
		return r.exists, r.problem
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

// GetLogs retrieves all log entries for the given level.  If level is AllEvents then all events will be retrieved
// todo: merge with GetEvents to clean up interfaces
func (c *Client) GetLogs(ctx context.Context, forEntity Meta, level EventLevel) ([]Event, bool, error) {
	result := make(chan getLogEventsReply, 1)
	select {
	case c.dataPlane <- &getLogEvents{
		replyTo: result,
		entity:  forEntity,
		level:   level,
	}:
	case <-ctx.Done():
		return nil, false, ctx.Err()
	}

	select {
	case out := <-result:
		return out.events, out.exists, out.problem
	case <-ctx.Done():
		return nil, false, ctx.Err()
	}
}

func (c *Client) GetEvents(ctx context.Context, forEntity Meta, level EventLevel) ([]Event, bool, error) {
	return c.GetLogs(ctx, forEntity, level)
}

type getLogEventsReply struct {
	problem error
	exists  bool
	events  []Event
}

type getLogEvents struct {
	replyTo chan getLogEventsReply
	entity  Meta
	level   EventLevel
}

func (g *getLogEvents) name() string {
	return "get-log-events"
}

func (g *getLogEvents) perform(ctx context.Context, c *Controller) {
	node, err := c.getNode(ctx, g.entity)
	if err != nil {
		g.replyTo <- getLogEventsReply{
			problem: err,
		}
		return
	}
	if node == nil {
		g.replyTo <- getLogEventsReply{
			exists: false,
		}
	}
	out := getLogEventsReply{
		exists: true,
	}
	if g.level == AllEvents {
		out.events = make([]Event, len(node.events))
		copy(out.events, node.events)
	} else {
		for _, l := range node.events {
			if l.Level < g.level {
				out.events = append(out.events, l)
			}
		}
	}
	select {
	case g.replyTo <- out:
	case <-ctx.Done():
	}
	close(g.replyTo) //todo: does not allow reuse on repaid changes
}
