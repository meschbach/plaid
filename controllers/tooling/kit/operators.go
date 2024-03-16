package kit

import (
	"context"
	"github.com/meschbach/plaid/resources"
)

type Operations[Spec any, Status any, State any] interface {
	//Create notifies the controller a new resource has been created and internal state should be created to reconcile
	//towards the specified state.
	Create(ctx context.Context, which resources.Meta, spec Spec, bridge Manager) (*State, error)
	//Update notifies the controller an update to the spec has occurred and internal state should be reconciled towards
	//that direction.
	Update(ctx context.Context, which resources.Meta, rt *State, s Spec) error
	//UpdateState asks the controller to investigate the environment for any changes in state.  Particularly useful when
	//claimed resources change or other cases when the spec has not changed.
	UpdateState(ctx context.Context, which resources.Meta, rt *State) error
	//Delete notifies the operations controller the spec for a specific resource has been deleted.
	Delete(ctx context.Context, which resources.Meta, rt *State) error
	//Status asks for the internal state to be represented as the externalized status.
	Status(ctx context.Context, rt *State) Status
}
