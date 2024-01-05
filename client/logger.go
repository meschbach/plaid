package client

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/internal/plaid/resources"
	"github.com/meschbach/plaid/ipc/grpc/logger"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
	"time"
)

// todo: merge with logdrain.LogEntry
type LogEntry struct {
	When       time.Time
	Text       string
	From       resources.Meta
	StreamName string
}

type LoggingDrainPump struct {
	ipc     logger.V1Client
	drainID int64
	offset  int64
	push    streams.Sink[LogEntry]
}

func (l *LoggingDrainPump) Serve(ctx context.Context) error {
	//establish watch
	registeredDrain, err := l.ipc.WatchDrain(ctx)
	if err != nil {
		return err
	}

	if err := registeredDrain.Send(&logger.WatchDrainRequest{
		DrainID: &l.drainID,
	}); err != nil {
		return err
	}

	for {
		_, err := registeredDrain.Recv()
		if err != nil {
			return err
		}
		readResult, err := l.ipc.ReadDrain(ctx, &logger.ReadDrainRequest{
			DrainID: l.drainID,
			Offset:  l.offset,
			Count:   32,
		})
		if err != nil {
			return err
		}
		l.offset = readResult.NextOffset

		for _, e := range readResult.Entries {
			if err := l.push.Write(ctx, LogEntry{
				When: e.When.AsTime(),
				Text: e.Text,
				From: resources.Meta{
					Type: resources.Type{
						Kind:    e.Origin.Kind,
						Version: e.Origin.Version,
					},
					Name: e.Origin.Name,
				},
				StreamName: e.Origin.Stream,
			}); err != nil {
				return err
			}
		}
	}
}

func PumpLogDrainEvents(ipc logger.V1Client, registration *logger.RegisterDrainReply, to streams.Sink[LogEntry]) *LoggingDrainPump {
	return &LoggingDrainPump{
		ipc:     ipc,
		drainID: registration.DrainID,
		offset:  registration.InitialOffset,
		push:    to,
	}
}

func NewTerminalLogger() *TerminalLoggerOutput {
	return &TerminalLoggerOutput{events: &streams.SinkEvents[LogEntry]{}}
}

type TerminalLoggerOutput struct {
	events *streams.SinkEvents[LogEntry]
}

func (t *TerminalLoggerOutput) Write(ctx context.Context, v LogEntry) error {
	fmt.Printf("[%s] %s:\t\t%s\n", v.From, v.StreamName, v.Text)
	return nil
}

func (t *TerminalLoggerOutput) Finish(ctx context.Context) error {
	return nil
}

func (t *TerminalLoggerOutput) SinkEvents() *streams.SinkEvents[LogEntry] {
	return t.events
}

func (t *TerminalLoggerOutput) Resume(ctx context.Context) error {
	return nil
}
