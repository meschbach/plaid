package client

import (
	"context"
	"github.com/meschbach/plaid/ipc/grpc/reswire"
	"github.com/meschbach/plaid/resources"
	"github.com/thejerf/suture/v4"
	"google.golang.org/grpc"
)

type Client interface {
	Create(ctx context.Context, ref resources.Meta, spec any) error
	Delete(ctx context.Context, ref resources.Meta) error
	Get(ctx context.Context, ref resources.Meta, spec any) (bool, error)
	GetStatus(ctx context.Context, ref resources.Meta, status any) (bool, error)
	GetEvents(ctx context.Context, ref resources.Meta, level resources.EventLevel) ([]resources.Event, error)
	List(ctx context.Context, kind resources.Type) ([]resources.Meta, error)
	Watcher(ctx context.Context) (Watcher, error)
}

type Watcher interface {
	OnType(ctx context.Context, kind resources.Type, consume resources.OnResourceChanged) (resources.WatchToken, error)
	OnResource(ctx context.Context, ref resources.Meta, consume resources.OnResourceChanged) (resources.WatchToken, error)
	Off(ctx context.Context, token resources.WatchToken) error
	Close(ctx context.Context) error
}

func New(transport *grpc.ClientConn, tree *suture.Supervisor) resources.System {
	wireClient := reswire.NewResourceControllerClient(transport)
	storageSupervisor := suture.NewSimple("wire.storage")
	tree.Add(storageSupervisor)
	storageWrapper := NewWireClientAdapter(storageSupervisor, wireClient)
	return &daemonSystem{
		d: &Daemon{
			grpcLayer:   transport,
			WireStorage: wireClient,
			Storage:     storageWrapper,
			LoggerV1:    nil,
			Tree:        storageSupervisor,
		},
	}
}
