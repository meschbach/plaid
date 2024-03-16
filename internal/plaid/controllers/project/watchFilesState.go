package project

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/controllers/filewatch"
	"github.com/meschbach/plaid/controllers/tooling"
	"github.com/meschbach/plaid/resources"
	"time"
)

type watchFilesState struct {
	basePath       string
	spec           bool
	lastFileChange *time.Time
	//restart indicates a change in the token has occurred and clients should propagate spec
	restart      bool
	restartToken string

	watcher tooling.Subresource[filewatch.Alpha1Status]
}

func (w *watchFilesState) updateSpec(spec *bool, defaultValue bool, baseDirectory string) {
	w.basePath = baseDirectory
	if spec == nil {
		w.spec = defaultValue
	} else {
		w.spec = *spec
	}
}

func (w *watchFilesState) reconcile(ctx context.Context, env tooling.Env) error {
	if !w.spec {
		return w.watcher.Delete(ctx, env)
	}

	var status filewatch.Alpha1Status
	step, err := w.watcher.Decide(ctx, env, &status)
	if err != nil {
		return err
	}
	switch step {
	case tooling.SubresourceCreate:
		ref := resources.Meta{
			Type: filewatch.Alpha1,
			Name: env.Subject.Name,
		}
		spec := filewatch.Alpha1Spec{AbsolutePath: w.basePath}
		return w.watcher.Create(ctx, env, ref, spec)
	case tooling.SubresourceExists:
		w.restart = false
		//waiting for a change
		if status.LastChange == w.lastFileChange {
			return nil
		}
		w.lastFileChange = status.LastChange
		w.restart = true
		w.restartToken = (*w.lastFileChange).Format(time.RFC3339)
	default:
		panic(fmt.Sprintf("unexpected subresource state %s\n", step))
	}
	return nil
}

func (w *watchFilesState) toStatus() *resources.Meta {
	if w.spec {
		return &w.watcher.Ref
	} else {
		return nil
	}
}
