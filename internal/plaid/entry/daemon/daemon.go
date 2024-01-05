// Package daemon provides Plaid as a daemonized system.
package daemon

import (
	"context"
	"errors"
	"fmt"
	"github.com/meschbach/plaid/internal/plaid/daemon"
	"github.com/meschbach/plaid/internal/plaid/ephemeral"
	"github.com/meschbach/plaid/internal/plaid/resources"
	"github.com/meschbach/go-junk-bucket/pkg/observability"
	"github.com/thejerf/suture/v4"
	"time"
)

type Config struct {
	//ProcessContext is the context which delineates the lifetime of the daemon
	ProcessContext context.Context
	//UnixSocketPath will launch a gRPC service of a Unix Domain Socket at the specified path.
	UnixSocketPath string
}

func DefaultConfig(procContext context.Context) Config {
	return Config{
		ProcessContext: procContext,
		UnixSocketPath: ephemeral.ResolvePlaidSocketPath(),
	}
}

func RunWithConfig(c Config) error {
	cfg := observability.DefaultConfig("plaid-daemon")
	component, err := cfg.Start(c.ProcessContext)
	if err != nil {
		return err
	}
	defer func() {
		shutdownContext, done := context.WithTimeout(context.Background(), 5*time.Second)
		defer done()
		if err := component.ShutdownGracefully(shutdownContext); err != nil {
			panic(err)
		}
	}()

	app := &daemonApp{
		Supervisor: *suture.NewSimple("daemons"),
		config:     c,
	}
	root := suture.NewSimple("root")
	root.Add(app)
	err = root.Serve(c.ProcessContext)
	if errors.Is(err, context.Canceled) {
		err = nil
	}
	return err
}

type daemonApp struct {
	suture.Supervisor
	config Config
}

func (d *daemonApp) Serve(ctx context.Context) error {
	fmt.Println("Starting Plaid daemon")
	registry := resources.NewController()
	d.Add(registry)

	bundled := newBundledService(registry)
	d.Add(bundled)

	grpcEndpoint := suture.NewSimple("grpc-endpoint")
	d.Add(grpcEndpoint)
	grpcEndpoint.Add(&daemon.GRPCService{
		Resources:     registry,
		Tree:          grpcEndpoint,
		LoggingConfig: bundled,
		Socket:        d.config.UnixSocketPath,
	})

	fmt.Println("Supervision tree configuration completed.")
	return d.Supervisor.Serve(ctx)
}
