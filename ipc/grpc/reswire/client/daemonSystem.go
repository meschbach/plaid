package client

import (
	"context"
	"github.com/meschbach/plaid/resources"
)

type daemonSystem struct {
	d *Daemon
}

func (d *daemonSystem) Storage(ctx context.Context) (resources.Storage, error) {
	return &daemonStorage{
		client: d.d.Storage,
		wire:   d.d.WireStorage,
	}, nil
}

func SystemFromDaemonV1(d *Daemon) resources.System {
	return &daemonSystem{d: d}
}
