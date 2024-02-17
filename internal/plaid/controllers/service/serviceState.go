package service

import (
	"github.com/meschbach/plaid/internal/plaid/controllers/dependencies"
	"github.com/meschbach/plaid/resources/operator"
)

type serviceState struct {
	bridge *operator.KindBridgeState

	token        string
	dependencies *dependencies.State
	build        builderState
	run          runState
	readiness    readinessProbeState
}
