package fileWatch

import (
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/operator"
	"time"
)

// watch represents a single resource for watching a specific point in the file system
type watch struct {
	meta resources.Meta
	//updateStatus is the interpreter to update the target resource
	updateStatus *alphaV1Interpreter
	//base is the prefix in the file system we are currently watching.  if empty string then nothing is being watched.
	base string
	//recursive
	recursive   bool
	watching    bool
	bridge      *operator.KindBridgeState
	lastUpdated *time.Time
}

func (w *watch) updateLastChange() {
	t := time.Now()
	w.lastUpdated = &t
}
