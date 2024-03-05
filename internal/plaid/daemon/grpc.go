package daemon

import (
	"context"
	"errors"
	"fmt"
	"github.com/meschbach/plaid/ipc/grpc/logger"
	"github.com/meschbach/plaid/ipc/grpc/reswire"
	"github.com/meschbach/plaid/resources"
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
	grpcLayer   *grpc.ClientConn
	WireStorage reswire.ResourceControllerClient
	Storage     Client
	LoggerV1    logger.V1Client
	Tree        *suture.Supervisor
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

	wireClient := reswire.NewResourceControllerClient(conn)
	resourceAdapter := NewWireClientAdapter(parent, wireClient)
	loggerV1Endpoint := logger.NewV1Client(conn)
	d := &Daemon{
		grpcLayer:   conn,
		WireStorage: wireClient,
		Storage:     resourceAdapter,
		LoggerV1:    loggerV1Endpoint,
		Tree:        parent,
	}
	return d, func() error {
		return d.Disconnect()
	}, nil
}

func typeToWire(t resources.Type) *reswire.Type {
	return &reswire.Type{
		Kind:    t.Kind,
		Version: t.Version,
	}
}

func metaToWire(ref resources.Meta) *reswire.Meta {
	return &reswire.Meta{
		Kind: typeToWire(ref.Type),
		Name: ref.Name,
	}
}

func externalizeEventLevel(l resources.EventLevel) reswire.EventLevel {
	switch l {
	case resources.AllEvents:
		return reswire.EventLevel_All
	case resources.Info:
		return reswire.EventLevel_Info
	case resources.Error:
		return reswire.EventLevel_Error
	default:
		panic(fmt.Sprintf("unhandled translation from %d", l))
	}
}

func internalizeEventLevel(l reswire.EventLevel) resources.EventLevel {
	switch l {
	case reswire.EventLevel_All:
		return resources.AllEvents
	case reswire.EventLevel_Error:
		return resources.Error
	case reswire.EventLevel_Info:
		return resources.Info
	default:
		panic(fmt.Sprintf("unhandled translation from %d", l))
	}
}

func internalizeOperation(op reswire.WatcherEventOut_Op) resources.ResourceChangedOperation {
	switch op {
	case reswire.WatcherEventOut_Created:
		return resources.CreatedEvent
	case reswire.WatcherEventOut_UpdatedStatus:
		return resources.StatusUpdated
	case reswire.WatcherEventOut_Deleted:
		return resources.DeletedEvent
	default:
		panic(fmt.Sprintf("unknown value %q", op.String()))
	}
}
