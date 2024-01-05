package up

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/meschbach/plaid/internal/plaid/daemon"
	"github.com/meschbach/plaid/internal/plaid/resources"
)

type ReportProgress[T any] struct {
	Prefix   string
	Core     *daemon.Daemon
	Of       resources.Meta
	OnChange func(ctx context.Context, status T) error
}

func (r *ReportProgress[T]) Watch(ctx context.Context) error {
	watch, err := r.Core.Storage.Watcher(ctx)
	if err != nil {
		return err
	}
	//todo: fix wire protocol to support closing
	//defer func() {
	//	if err := watch.Close(ctx); err != nil {
	//		panic(err)
	//	}x
	//}()

	report := func(ctx context.Context) error {
		var status T
		exists, err := r.Core.Storage.GetStatus(ctx, r.Of, &status)
		if err != nil || !exists {
			return err
		}
		j, err := json.Marshal(status)
		if err != nil {
			return err
		}
		fmt.Printf("%s\t\t%s\n", r.Prefix, j)

		return r.OnChange(ctx, status)
	}

	token, err := watch.OnResource(ctx, r.Of, func(ctx context.Context, changed resources.ResourceChanged) error {
		return report(ctx)
	})
	if err != nil {
		return err
	}
	defer func() {
		if err := watch.Off(ctx, token); err != nil {
			panic(err)
		}
	}()

	if err := report(ctx); err != nil {
		return err
	}
	return nil
}
