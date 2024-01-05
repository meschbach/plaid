package logdrain

import (
	"context"
	"errors"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
)

type LeakyBufferEntry[E any] struct {
	Leaked uint
	Entry  E
}

type leakyBufferState uint8

const (
	bufferPaused leakyBufferState = iota
	bufferFinished
)

// LeakyBuffer is a Sink and Source which will drop the oldest message to prevent buffer overflows.  As a source each
// returned item will return the count of dropped items ahead of it.
type LeakyBuffer[E any] struct {
	state        leakyBufferState
	limit        uint
	sinkEvents   *streams.SinkEvents[E]
	sourceEvents *streams.SourceEvents[LeakyBufferEntry[E]]
	buffer       []LeakyBufferEntry[E]
}

func (l *LeakyBuffer[E]) Write(ctx context.Context, v E) error {
	switch l.state {
	case bufferFinished:
		return streams.Done
	}

	l.buffer = append(l.buffer, LeakyBufferEntry[E]{
		Leaked: 0,
		Entry:  v,
	})
	//Buffer is full, drop the oldest message.
	if len(l.buffer) > int(l.limit) {
		newDropCount := l.buffer[0].Leaked + 1
		l.buffer = l.buffer[1:]
		l.buffer[0].Leaked = newDropCount
	}
	return nil
}

func (l *LeakyBuffer[E]) Finish(ctx context.Context) error {
	switch l.state {
	case bufferFinished:
		return nil
	}
	l.state = bufferFinished
	return nil
}

func (l *LeakyBuffer[E]) SinkEvents() *streams.SinkEvents[E] {
	return l.sinkEvents
}

// Resume instructs the Sink to resume draining internal buffers and accept writes again.
func (l *LeakyBuffer[E]) Resume(ctx context.Context) error {
	return errors.New("todo")
}

func (l *LeakyBuffer[E]) Pause(ctx context.Context) error {
	return errors.New("todo")
}

func (l *LeakyBuffer[E]) SourceEvents() *streams.SourceEvents[LeakyBufferEntry[E]] {
	return l.sourceEvents
}
func (l *LeakyBuffer[E]) ReadSlice(ctx context.Context, to []LeakyBufferEntry[E]) (int, error) {
	count := copy(to, l.buffer)
	l.buffer = l.buffer[count:]
	return count, nil
}

func NewLeakyBuffer[E any](limit uint) *LeakyBuffer[E] {
	return &LeakyBuffer[E]{
		sinkEvents:   &streams.SinkEvents[E]{},
		sourceEvents: &streams.SourceEvents[LeakyBufferEntry[E]]{},
		limit:        limit,
	}
}
