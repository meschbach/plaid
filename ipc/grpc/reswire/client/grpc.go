package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/meschbach/plaid/ipc/grpc/logger"
	"github.com/meschbach/plaid/ipc/grpc/reswire"
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
