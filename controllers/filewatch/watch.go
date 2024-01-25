package filewatch

import (
	"context"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/operator"
	"time"
)

// watch represents a single resource for watching a specific point in the file system
type watch struct {
	meta resources.Meta
	//base is the prefix in the file system we are currently watching.  if empty string then nothing is being watched.
	base string
	//recursive
	watching    bool
	bridge      *operator.KindBridgeState
	lastUpdated *time.Time
}

func (w *watch) flushStatus(ctx context.Context) error {
	return w.bridge.OnResourceChange(ctx, w.meta)
}
