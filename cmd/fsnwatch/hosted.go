package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/meschbach/go-junk-bucket/pkg/observability"
	"github.com/meschbach/plaid/controllers/filewatch/fsn"
	"github.com/meschbach/plaid/internal/plaid/ephemeral"
	"github.com/meschbach/plaid/ipc/grpc/reswire/client"
	"github.com/spf13/cobra"
	"github.com/thejerf/suture/v4"
	"golang.org/x/sys/unix"
	"os"
	"os/signal"
)

type plaidOpts struct {
	Service string
}

func withPlaidCobra(cfg *plaidOpts, perform func(ctx context.Context, plaid *client.Daemon, tree *suture.Supervisor, args []string) error) func(command *cobra.Command, args []string) error {
	return func(command *cobra.Command, args []string) error {
		cmdCtx := command.Context()
		o11yCfg := observability.DefaultConfig("fsn-watch")
		o11y, err := o11yCfg.Start(cmdCtx)
		if err != nil {
			return err
		}
		defer o11y.ShutdownGracefully(cmdCtx)

		procContext, done := signal.NotifyContext(cmdCtx, unix.SIGINT, unix.SIGTERM, unix.SIGHUP)
		defer done()

		//establish connection
		operationsTree := suture.NewSimple("root")
		wireClient, onDisconnect, err := client.DialClient(procContext, cfg.Service, operationsTree)
		if err != nil {
			if _, err := fmt.Fprintln(os.Stderr, err.Error()); err != nil {
				panic(err)
			}
			return nil
		}
		defer onDisconnect()

		return perform(procContext, wireClient, operationsTree, args)
	}
}

func runHosted(ctx context.Context, plaid *client.Daemon, tree *suture.Supervisor, args []string) error {
	sys := client.SystemFromDaemonV1(plaid)

	fmt.Println("Connected, listing")
	fsnTree := fsn.NewFileWatchSystem(sys)
	tree.Add(fsnTree)
	result := tree.Serve(ctx)
	fmt.Println("Shutting down.")
	if errors.Is(result, context.Canceled) {
		result = nil
	}
	return result
}

func hostedCommand() *cobra.Command {
	hostedOpts := &plaidOpts{}
	cmd := &cobra.Command{
		Use:   "hosted",
		Short: "connects to a plaid resource service and provides file watcher resources",
		RunE:  withPlaidCobra(hostedOpts, runHosted),
	}
	cmd.Flags().StringVar(&hostedOpts.Service, "plaid-address", ephemeral.ResolvePlaidSocketPath(), "Address for connecting to plaid")

	return cmd
}
