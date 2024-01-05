package daemon

import (
	"context"
	"errors"
	"fmt"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/daemon/wire"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"git.meschbach.com/mee/platform.git/plaid/ipc/grpc/logger"
	"github.com/thejerf/suture/v4"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"time"
)

type UnableToConnect struct {
	socket string
}

func (u *UnableToConnect) Error() string {
	return fmt.Sprintf("unable to connect to %q", u.socket)
}

type Daemon struct {
	grpcLayer *grpc.ClientConn
	Storage   Client
	LoggerV1  logger.V1Client
	Tree      *suture.Supervisor
}

func (d *Daemon) Disconnect() error {
	return d.grpcLayer.Close()
}

func DialClient(ctx context.Context, address string, parent *suture.Supervisor) (*Daemon, func() error, error) {
	var (
		instrumentation = otelgrpc.NewClientHandler()
		credentials     = insecure.NewCredentials() // No SSL/TLS
		dialer          = func(ctx context.Context, addr string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", addr)
		}
		options = []grpc.DialOption{
			grpc.WithStatsHandler(instrumentation),
			grpc.WithTransportCredentials(credentials),
			grpc.WithBlock(),
			grpc.WithContextDialer(dialer),
		}
	)

	var dialTimeout context.Context
	var dailDone func()
	var conn *grpc.ClientConn
	for {
		dialTimeout, dailDone = context.WithTimeout(ctx, 1*time.Second)
		var err error
		conn, err = grpc.DialContext(dialTimeout, address, options...)
		if err == nil {
			break
		} else {
			if errors.Is(err, context.DeadlineExceeded) {
				return nil, func() error {
					dailDone()
					return nil
				}, &UnableToConnect{socket: address}
			}
			return nil, func() error {
				dailDone()
				return nil
			}, err
		}
	}
	defer dailDone()

	wireClient := wire.NewResourceControllerClient(conn)
	resourceAdapter := NewWireClientAdapter(parent, wireClient)
	loggerV1Endpoint := logger.NewV1Client(conn)
	d := &Daemon{
		grpcLayer: conn,
		Storage:   resourceAdapter,
		LoggerV1:  loggerV1Endpoint,
		Tree:      parent,
	}
	return d, func() error {
		return d.Disconnect()
	}, nil
}

func typeToWire(t resources.Type) *wire.Type {
	return &wire.Type{
		Kind:    t.Kind,
		Version: t.Version,
	}
}

func metaToWire(ref resources.Meta) *wire.Meta {
	return &wire.Meta{
		Kind: typeToWire(ref.Type),
		Name: ref.Name,
	}
}

func externalizeEventLevel(l resources.EventLevel) wire.EventLevel {
	switch l {
	case resources.AllEvents:
		return wire.EventLevel_All
	case resources.Info:
		return wire.EventLevel_Info
	case resources.Error:
		return wire.EventLevel_Error
	default:
		panic(fmt.Sprintf("unhandled translation from %d", l))
	}
}

func internalizeEventLevel(l wire.EventLevel) resources.EventLevel {
	switch l {
	case wire.EventLevel_All:
		return resources.AllEvents
	case wire.EventLevel_Error:
		return resources.Error
	case wire.EventLevel_Info:
		return resources.Info
	default:
		panic(fmt.Sprintf("unhandled translation from %d", l))
	}
}

func internalizeOperation(op wire.WatcherEventOut_Op) resources.ResourceChangedOperation {
	switch op {
	case wire.WatcherEventOut_Created:
		return resources.CreatedEvent
	case wire.WatcherEventOut_UpdatedStatus:
		return resources.StatusUpdated
	default:
		panic(fmt.Sprintf("unknown value %q", op.String()))
	}
}
