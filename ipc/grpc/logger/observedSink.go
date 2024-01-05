package logger

import (
	"context"
	"errors"
	"github.com/meschbach/go-junk-bucket/pkg/emitter"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
)

type observedSink[T any] struct {
	writtenEvent emitter.Dispatcher[streams.Sink[T]]
	target       streams.Sink[T]
}

func (o *observedSink[T]) Write(ctx context.Context, element T) error {
	err := o.target.Write(ctx, element)
	if err == nil || (errors.Is(err, streams.Full)) {
		if eventError := o.writtenEvent.Emit(ctx, o.target); eventError != nil {
			return errors.Join(err, eventError)
		}
	}
	return err
}

func (o *observedSink[T]) Finish(ctx context.Context) error {
	return o.target.Finish(ctx)
}

func (o *observedSink[T]) SinkEvents() *streams.SinkEvents[T] {
	return o.target.SinkEvents()
}

func (o *observedSink[T]) Resume(ctx context.Context) error {
	return o.Resume(ctx)
}
