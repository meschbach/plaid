package get

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/meschbach/plaid/internal/plaid/daemon"
	"github.com/meschbach/plaid/resources"
	"os"
)

func Perform(ctx context.Context, client *daemon.Daemon, options Options) error {
	ref := resources.Meta{
		Type: resources.Type{
			Kind:    options.Kind,
			Version: options.Version,
		},
		Name: options.Resource,
	}

	var out json.RawMessage
	exists, err := client.Storage.Get(ctx, ref, &out)
	if err != nil {
		return err
	}

	if !exists {
		fmt.Println("Resource does not exist")
		return nil
	}

	var status json.RawMessage
	_, err = client.Storage.GetStatus(ctx, ref, &status)
	if err != nil {
		return err
	}

	if options.PrettyJSON {
		type output struct {
			Spec   json.RawMessage `json:"spec"`
			Status json.RawMessage `json:"status"`
		}

		o := output{
			Spec:   out,
			Status: status,
		}
		data, err := json.MarshalIndent(o, "", "\t")
		if err != nil {
			return err
		}
		if _, err := os.Stdout.Write(data); err != nil {
			return err
		}
	} else {
		fmt.Printf("Spec: %s\n", out)
		fmt.Printf("Status: %s\n", status)
	}
	return nil
}
