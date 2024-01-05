package main

import (
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
			ctx, _ := signal.NotifyContext(cmd.Context(), unix.SIGTERM)
			cfg := daemon.DefaultConfig(ctx)
			return daemon.RunWithConfig(cfg)
		},
	})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
