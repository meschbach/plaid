package service

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/internal/plaid/controllers/logdrain"
	"github.com/meschbach/plaid/ipc/grpc/logger"
	"github.com/meschbach/plaid/ipc/grpc/reswire"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/service/logging"
	"github.com/thejerf/suture/v4"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
)

type LoggingConfig interface {
	LogDrainConfig(ctx context.Context) *logdrain.ServiceConfig
}

type GRPCService struct {
	Resources     *resources.Controller
	Tree          *suture.Supervisor
	LoggingConfig LoggingConfig
	Socket        string
}

func (g *GRPCService) Serve(ctx context.Context) error {
	c := g.Resources.Client()
	service := &ResourceService{
		client: c,
	}
	loggingConfig := g.LoggingConfig.LogDrainConfig(ctx)
	v1Logging := logging.NewV1GPRCBridge(g.Tree, g.Resources, loggingConfig)

	var sockAddr string
	if len(g.Socket) == 0 {
		sockAddr = "/tmp/plaid-dev.socket"
	} else {
		sockAddr = g.Socket
	}
	fmt.Printf("[gRPC] starting at %q\n", sockAddr)

	if _, err := os.Stat(sockAddr); !os.IsNotExist(err) {
		if err := os.RemoveAll(sockAddr); err != nil {
			//todo: log?
			log.Fatal(err)
		}
	}
	protocol := "unix"

	listener, err := net.Listen(protocol, sockAddr)
	if err != nil {
		log.Fatal(err)
	}

	serviceInstrumentation := otelgrpc.NewServerHandler()
	serviceOptions := []grpc.ServerOption{
		grpc.StatsHandler(serviceInstrumentation),
	}
	server := grpc.NewServer(serviceOptions...)
	reswire.RegisterResourceControllerServer(server, service)
	logger.RegisterV1Server(server, v1Logging)
	problem := make(chan error, 1)
	go func() {
		problem <- server.Serve(listener)
	}()

	select {
	case err := <-problem:
		return err
	case <-ctx.Done():
		server.Stop()
		return ctx.Err()
	}
}
