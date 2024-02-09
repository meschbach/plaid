package operator

import (
	"context"
	"github.com/meschbach/plaid/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var capturingKind = resources.Type{
	Kind:    "test.meschbach.com/capturing-kind",
	Version: "1",
}

type capturingKindState struct {
	which   resources.Meta
	created int
	updated int
	deleted int
}

type capturingKindSpec struct {
}

type capturingKindStatus struct {
}

type capturingKindController struct {
	created bool
	last    *capturingKindState
}

func (c *capturingKindController) Create(ctx context.Context, which resources.Meta, spec capturingKindSpec, bridge *KindBridgeState) (*capturingKindState, capturingKindStatus, error) {
	state := &capturingKindState{which: which, created: 1}
	c.created = true
	c.last = state
	status, err := c.Update(ctx, which, state, spec)
	return state, status, err
}
func (c *capturingKindController) Update(ctx context.Context, which resources.Meta, rt *capturingKindState, s capturingKindSpec) (capturingKindStatus, error) {
	rt.updated++
	return capturingKindStatus{}, nil
}
func (c *capturingKindController) Delete(ctx context.Context, which resources.Meta, rt *capturingKindState) error {
	rt.deleted++
	return nil
}

func TestKindBridge(t *testing.T) {
	t.Run("Given a new resources", func(t *testing.T) {
		root, onDone := context.WithCancel(context.Background())
		t.Cleanup(onDone)
		core := resources.WithTestSubsystem(t, root)
		controller := &capturingKindController{}

		k := NewKindBridge[capturingKindSpec, capturingKindStatus, capturingKindState](capturingKind, controller)
		_, err := k.Setup(root, core.Store)
		require.NoError(t, err)

		ref := resources.Meta{
			Type: capturingKind,
			Name: "first-subject",
		}
		require.NoError(t, core.Store.Create(root, ref, capturingKindSpec{}))

		var state *capturingKindState
		for !controller.created {
			select {
			case <-root.Done():
				require.NoError(t, root.Err())
				return
			case event := <-k.observer.Feed:
				require.NoError(t, k.observer.Digest(root, event))
				state = controller.last
			}
		}

		t.Run("When the resource is deleted", func(t *testing.T) {
			existed, err := core.Store.Delete(root, ref)
			require.True(t, existed, "target resource should still exist")
			require.NoError(t, err)
			for state.deleted == 0 {
				select {
				case <-root.Done():
					require.NoError(t, root.Err())
					return
				case event := <-k.observer.Feed:
					require.NoError(t, k.observer.Digest(root, event))
				}
			}
			assert.Less(t, 0, state.deleted, "then the resource is deleted")
		})
	})
}
