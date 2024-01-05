package buildrun

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/controllers/dependencies"
	"github.com/meschbach/plaid/internal/plaid/resources"
	"github.com/meschbach/plaid/internal/plaid/resources/operator"
)

const procAnnotationRole = Kind + ":role"
const procAnnotationToken = Kind + ":restart-token"

const procAnnotationRoleBuilder = "builder"
const procAnnotationRoleProc = "proc"

type state struct {
	bridge   *operator.KindBridgeState
	builder  builderState
	proc     runState
	requires dependencies.State
}

type stateEnv struct {
	object       resources.Meta
	rpc          *resources.Client
	watcher      *resources.ClientWatcher
	reconcile    func(ctx context.Context) error
	restartToken string
}
