package daemon

import (
	"context"
	"github.com/meschbach/plaid/ipc/grpc/reswire"
	"github.com/thejerf/suture/v4"
	"io"
)

type serviceWatcherInputPump struct {
	stream reswire.ResourceController_WatcherServer
	events chan *reswire.WatcherEventIn
}

func (s *serviceWatcherInputPump) Serve(ctx context.Context) error {
	for {
		event, err := s.stream.Recv()
		if err != nil {
			if err == io.EOF {
				close(s.events)
				return suture.ErrDoNotRestart
			} else {
				return err
			}
		}
		select {
		case s.events <- event:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
