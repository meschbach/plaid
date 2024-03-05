package main

import (
	"fmt"
	"github.com/meschbach/plaid/internal/plaid/entry/daemon"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"os"
	"os/signal"
)

func main() {
	rootCmd := &cobra.Command{
		Use:          "plaid-daemon",
		Short:        "Plaid daemon",
		Long:         "Platform, Library, and Application implement develop for rapid development",
		SilenceUsage: true,
	}
	rootCmd.AddCommand(&cobra.Command{
		Use: "run",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, _ := signal.NotifyContext(cmd.Context(), unix.SIGTERM, unix.SIGINT)
			go func() {
				<-ctx.Done()
				fmt.Println("[pliadd] Shutting down.")
			}()
			cfg := daemon.DefaultConfig(ctx)
			return daemon.RunWithConfig(cfg)
		},
	})

	if err := rootCmd.Execute(); err != nil {
		_, e := fmt.Fprintf(os.Stderr, "Failed to run with %s\n", err.Error())
		if e != nil {
			panic(e)
		}
	}
}
