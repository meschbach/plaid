package main

import (
	"context"
	"github.com/meschbach/plaid/client"
	client2 "github.com/meschbach/plaid/ipc/grpc/reswire/client"
	"github.com/meschbach/plaid/resources"
	"github.com/spf13/cobra"
)

func deleteCommand(rt *client.Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <kind> <version> <name>",
		Short: "deletes the specific resource in question",
		Args:  cobra.ExactArgs(3),
		RunE: runCommand("delete", rt, func(ctx context.Context, rt *client.Runtime, client *client2.Daemon, args []string) error {
			kind := resources.Type{Kind: args[0], Version: args[1]}
			ref := resources.Meta{
				Type: kind,
				Name: args[2],
			}

			err := client.Storage.Delete(ctx, ref)
			return err
		}),
	}
}
