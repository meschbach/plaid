package kit

import "context"

type Manager interface {
	//Update will prompt an attempt reconciliation as though there was an update to the spec.
	Update(ctx context.Context) error
	//UpdateState prompts the operations to investigate the internal state of the system for an update
	UpdateState(ctx context.Context) error
	//UpdateStatus will inquire about the internal cached state of the resource.
	UpdateStatus(ctx context.Context) error
}
