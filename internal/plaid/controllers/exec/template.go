package exec

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/resources"
	"os"
)

type TemplateAlpha1Spec struct {
	Command    string `json:"command"`
	WorkingDir string `json:"wd"`
}

func (t TemplateAlpha1Spec) CreateResource(ctx context.Context, client *resources.Client, claimer resources.Meta, annotations map[string]string, watcher *resources.ClientWatcher, consumer resources.OnResourceChanged) (resources.Meta, resources.WatchToken, error) {
	var token resources.WatchToken
	var wd string
	if t.WorkingDir != "" {
		wd = t.WorkingDir
	} else {
		current, err := os.Getwd()
		if err != nil {
			return resources.Meta{}, token, err
		}
		wd = current
	}

	//fill out template
	result := InvocationAlphaV1Spec{
		Exec:       t.Command,
		WorkingDir: wd,
	}

	//generate-id
	id := resources.GenSuffix(8)
	name := claimer.Name + "-" + id
	instanceRef := resources.Meta{Type: InvocationAlphaV1Type, Name: name}
	if watcher != nil {
		var err error
		token, err = watcher.OnResourceStatusChanged(ctx, instanceRef, consumer)
		if err != nil {
			return instanceRef, token, err
		}
	}

	//dispatch
	problem := client.Create(ctx, instanceRef, result, resources.ClaimedBy(claimer), resources.WithAnnotations(annotations))
	return instanceRef, token, problem
}
