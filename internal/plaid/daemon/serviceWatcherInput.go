package daemon

import (
	"context"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/daemon/wire"
	"github.com/thejerf/suture/v4"
	"io"
)

type serviceWatcherInputPump struct {
	stream wire.ResourceController_WatcherServer
	events chan *wire.WatcherEventIn
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
