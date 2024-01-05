package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/daemon"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
)

func Get(ctx context.Context, client *daemon.Daemon, kind string, version string, name string) error {
	ref := resources.Meta{
		Type: resources.Type{
			Kind:    kind,
			Version: version,
		},
		Name: name,
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

	fmt.Printf("Spec: %s\n", out)
	fmt.Printf("Status: %s\n", status)
	return nil
}
