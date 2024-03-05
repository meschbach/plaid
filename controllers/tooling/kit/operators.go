package kit

import (
	"context"
	"github.com/meschbach/plaid/resources"
)

type Operations[Spec any, Status any, State any] interface {
	Create(ctx context.Context, which resources.Meta, spec Spec, bridge Manager) (*State, error)
	Update(ctx context.Context, which resources.Meta, rt *State, s Spec) error
	Delete(ctx context.Context, which resources.Meta, rt *State) error
	Status(ctx context.Context, rt *State) Status
}
