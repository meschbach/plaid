package probes

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/httpProbe"
	"github.com/meschbach/plaid/resources"
)

type httpAdapter struct {
	state *httpProbe.TemplateState
}

func (h *httpAdapter) Reconcile(ctx context.Context, storage resources.Storage) error {
	return h.state.Reconcile(ctx, storage)
}
func (h *httpAdapter) Ready() bool {
	return h.state.Ready()
}

func (h *httpAdapter) Status() TemplateAlpha1Status {
	status := h.state.Status()
	return TemplateAlpha1Status{
		Ref:   status.Ref,
		Ready: status.Ready,
	}
}
