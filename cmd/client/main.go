package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/meschbach/go-junk-bucket/pkg/observability"
	"github.com/meschbach/plaid/client"
	"github.com/meschbach/plaid/client/get"
	"github.com/meschbach/plaid/internal/plaid/entry/client/usecase"
	"github.com/meschbach/plaid/internal/plaid/ephemeral"
	client2 "github.com/meschbach/plaid/ipc/grpc/reswire/client"
	"github.com/meschbach/plaid/resources"
	"github.com/spf13/cobra"
	"github.com/thejerf/suture/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/sys/unix"
	"os"
	"os/signal"
	"runtime/pprof"
	"time"
)

func main() {
	rt := &client.Runtime{ExitCode: 0}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, unix.SIGUSR1)
		<-c
		pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
	}()
	rootCmd := &cobra.Command{
		Use:          "plaid-client",
		Short:        "Plaid client",
		Long:         "Platform, Library, and Application implement develop for rapid development",
		SilenceUsage: true,
	}
	rootCmd.AddCommand(deleteCommand(rt))
	rootCmd.AddCommand(getCommand(rt))
	rootCmd.AddCommand(listCommand(rt))
	rootCmd.AddCommand(upCommand(rt))
	rootCmd.AddCommand(createCommand(rt))

	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
	os.Exit(rt.ExitCode)
}

func listCommand(rt *client.Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "list <kind> <version>",
		Short: "lists resources",
		Args:  cobra.ExactArgs(2),
		RunE: runCommand("list", rt, func(ctx context.Context, rt *client.Runtime, client *client2.Daemon, args []string) error {
			matching, err := client.Storage.List(ctx, resources.Type{
				Kind:    args[0],
				Version: args[1],
			})
			if err != nil {
				return err
			}
			for _, m := range matching {
				fmt.Printf("%s/%s:\t%s\n", m.Type.Kind, m.Type.Version, m.Name)
			}
			return nil
		}),
	}
}

func upCommand(rt *client.Runtime) *cobra.Command {
	opt := usecase.UpOptions{}
	cmd := &cobra.Command{
		Use:   "up",
		Short: "launches a manifest and waits for it to be 'complete'",
		Args:  cobra.ExactArgs(0),
		RunE: runCommand("up", rt, func(ctx context.Context, rt *client.Runtime, client *client2.Daemon, args []string) error {
			return usecase.Up(ctx, client, rt, opt)
		}),
	}
	cmd.Flags().BoolVarP(&opt.ReportUpdates, "report-progress", "p", false, "Reports status updates as they occur")
	cmd.Flags().BoolVarP(&opt.DeleteOnCompletion, "delete-on-completion", "d", false, "Deletes project on completion")
	return cmd
}

func getCommand(rt *client.Runtime) *cobra.Command {
	opts := &get.Options{}
	cmd := &cobra.Command{
		Use:   "get <kind> <version> <name>",
		Short: "Retrieves a resource",
		Args:  cobra.ExactArgs(3),
		RunE: runCommand("get", rt, func(ctx context.Context, rt *client.Runtime, client *client2.Daemon, args []string) error {
			opts.Kind = args[0]
			opts.Version = args[1]
			opts.Resource = args[2]
			return get.Perform(ctx, client, *opts)
		}),
	}
	f := cmd.Flags()
	f.BoolVarP(&opts.PrettyJSON, "pretty-json", "p", false, "pretty print JSON")
	return cmd
}

func runCommand(op string, rt *client.Runtime, fn func(ctx context.Context, rt *client.Runtime, d *client2.Daemon, args []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx, done := context.WithCancel(cmd.Context())
		defer done()

		o11yConfig := observability.DefaultConfig("plaid-control:" + cmd.Name())
		o11yConfig.Silent = true
		o11y, err := o11yConfig.Start(ctx)
		if err != nil {
			return err
		}
		defer func() {
			shutdown, done := context.WithTimeout(context.Background(), 10*time.Second)
			defer done()
			if err := o11y.ShutdownGracefully(shutdown); err != nil {
				panic(err)
			}
		}()

		tree := suture.NewSimple("app")
		app := &clientEnv{
			op: op,
			perform: func(ctx context.Context, rt *client.Runtime, client *client2.Daemon) error {
				return fn(ctx, rt, client, args)
			},
			pool: tree,
			done: done,
			rt:   rt,
		}
		tree.Add(app)

		err = tree.Serve(ctx)
		if errors.Is(err, suture.ErrTerminateSupervisorTree) || errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	}
}

type OnConnected func(ctx context.Context, rt *client.Runtime, daemon *client2.Daemon) error
type clientEnv struct {
	op      string
	pool    *suture.Supervisor
	perform OnConnected
	done    func()
	rt      *client.Runtime
}

func (c *clientEnv) Serve(serviceContext context.Context) error {
	ctx, span := tracer.Start(serviceContext, c.op)
	defer span.End()
	socketPath := ephemeral.ResolvePlaidSocketPath()

	client, clientDone, err := client2.DialClient(ctx, socketPath, c.pool)
	if err != nil {
		return err
	}
	defer clientDone()

	err = c.perform(ctx, c.rt, client)
	if err != nil {
		if errors.Is(err, suture.ErrDoNotRestart) || errors.Is(err, suture.ErrTerminateSupervisorTree) {
			//fall through for normal termination
		} else {
			span.SetStatus(codes.Error, "failed to perform requested operation")
			span.RecordError(err)
			return err
		}
	}
	c.done()
	return nil
}

var tracer = otel.Tracer("plaid-control")
