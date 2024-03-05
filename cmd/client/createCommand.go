package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/meschbach/go-junk-bucket/pkg/files"
	"github.com/meschbach/plaid/client"
	"github.com/meschbach/plaid/internal/plaid/daemon"
	"github.com/meschbach/plaid/internal/plaid/daemon/wire"
	"github.com/meschbach/plaid/ipc/grpc/reswire"
	"github.com/meschbach/plaid/resources"
	"github.com/spf13/cobra"
	"os"
)

type Manifest struct {
	Meta struct {
		Name resources.Meta `json:"name"`
	} `json:"meta"`
	Spec json.RawMessage `json:"spec"`
}

func createCommand(rt *client.Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "create <manifest>",
		Short: "Creates a manifest",
		Args:  cobra.ExactArgs(1),
		RunE: runCommand("delete", rt, func(ctx context.Context, rt *client.Runtime, client *daemon.Daemon, args []string) error {
			fileName := args[0]
			manifest := Manifest{}
			if err := files.ParseJSONFile(fileName, &manifest); err != nil {
				_, writeError := fmt.Fprintf(os.Stderr, "Failed ot parse JSON manifest %s: %s\n", fileName, err.Error())
				if writeError != nil {
					panic(writeError)
				}
				return nil
			}

			_, err := client.WireStorage.Create(ctx, &wire.CreateResourceIn{
				Target: reswire.MetaToWire(manifest.Meta.Name),
				Spec:   manifest.Spec,
			})
			if err != nil {
				_, writeError := fmt.Fprintf(os.Stderr, "Failed to create resource: %s\n", err.Error())
				if writeError != nil {
					panic(writeError)
				}
				return nil
			}
			fmt.Printf("Created %s\n", manifest)
			return nil
		}),
	}
}
