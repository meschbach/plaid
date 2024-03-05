package resbridge

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/daemon"
	"github.com/meschbach/plaid/resources"
)

type daemonSystem struct {
	d *daemon.Daemon
}

func (d *daemonSystem) Storage(ctx context.Context) (resources.Storage, error) {
	return &daemonStorage{
		client: d.d.Storage,
		wire:   d.d.WireStorage,
	}, nil
}

func SystemFromDaemonV1(d *daemon.Daemon) resources.System {
	return &daemonSystem{d: d}
}
