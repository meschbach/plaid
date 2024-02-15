package service

import (
	"github.com/meschbach/plaid/internal/plaid/controllers/dependencies"
	"github.com/meschbach/plaid/resources/operator"
)

type serviceState struct {
	bridge *operator.KindBridgeState

	dependencies *dependencies.State
	build        builderState
	run          runState
	readiness    readinessProbeState
}
