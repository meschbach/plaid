package httpProbe

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/resources"
)

type TemplateAlpha1 struct {
	Port uint16 `json:"port"`
	Path string `json:"path"`
}

func (t *TemplateAlpha1) Instantiate(ctx context.Context, storage *resources.Client, claimedBy resources.Meta, watch *resources.ClientWatcher, onChange func(ctx context.Context) error) (*TemplateState, error) {
	ref := resources.Meta{
		Type: AlphaV1Type,
		Name: fmt.Sprintf("%s-port-%d", claimedBy.Name, t.Port),
	}
	if err := storage.Create(ctx, ref, AlphaV1Spec{
		Enabled:  true,
		Host:     "localhost",
		Port:     t.Port,
		Resource: t.Path,
	}, resources.ClaimedBy(claimedBy)); err != nil {
		return nil, err
	}

	state := &TemplateState{ref: ref, ready: false}
	token, err := watch.OnResource(ctx, ref, func(ctx context.Context, changed resources.ResourceChanged) error {
		switch changed.Operation {
		case resources.StatusUpdated:
			return onChange(ctx)
		default:
		}
		return nil
	})
	if err != nil { //todo: dispose of resource
		return nil, err
	}
	state.token = token
	return state, nil
}

type TemplateState struct {
	ref   resources.Meta
	token resources.WatchToken
	ready bool
}

func (t *TemplateState) Reconcile(ctx context.Context, storage *resources.Client) error {
	var status AlphaV1Status
	exists, err := storage.GetStatus(ctx, t.ref, &status)
	if !exists || err != nil {
		t.ready = false
		return err
	}
	if t.ready != status.Ready {
		t.ready = status.Ready
		//todo: dispatch an update event
	}
	return nil
}

func (t *TemplateState) Ready() bool {
	return t.ready
}
