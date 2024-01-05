package logdrain

import (
	"context"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/operator"
)

const (
	Kind   = "plaid.meschbach.com/logs"
	Alpha1 = "alpha1"
)

var AlphaV1Type = resources.Type{
	Kind:    Kind,
	Version: Alpha1,
}

type Alpha1Spec struct {
	//Source is the target process to receive logging messages from
	Source AlphaV1SourceSpec `json:"source"`
	Drain  AlphaV1DrainSpec  `json:"drain"`
}

type AlphaV1SourceSpec struct {
	Ref    resources.Meta
	Stream string
}

type AlphaV1DrainSpec struct {
	Ref    resources.Meta
	Stream string
}

const (
	Unknown     = "unknown"
	Connected   = "connected"
	Unconnected = "unconnected"
)

type Alpha1Status struct {
	Pipe string `json:"pipe"`
}

type alphaV1 struct {
	c *Controller
}

func (a *alphaV1) Create(ctx context.Context, which resources.Meta, spec Alpha1Spec, bridge *operator.KindBridgeState) (*alpha1State, Alpha1Status, error) {
	state := &alpha1State{
		bridge: bridge,
		self:   which,
	}
	return state, state.toStatus(), nil
}

func (a *alphaV1) Update(ctx context.Context, which resources.Meta, rt *alpha1State, s Alpha1Spec) (Alpha1Status, error) {
	return rt.toStatus(), nil
}
