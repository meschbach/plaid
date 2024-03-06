package filewatch

import (
	"context"
	"github.com/meschbach/plaid/controllers/tooling/kit"
	"github.com/meschbach/plaid/resources"
	"time"
)

// watch represents a single resource for watching a specific point in the file system
type watch struct {
	meta resources.Meta
	//base is the prefix in the file system we are currently watching.  if empty string then nothing is being watched.
	base string
	//recursive
	watching    bool
	bridge      kit.Manager
	lastUpdated *time.Time
}

func (w *watch) flushStatus(ctx context.Context) error {
	return w.bridge.UpdateStatus(ctx)
}

func (w *watch) asStatus() Alpha1Status {
	var to Alpha1Status
	to.LastChange = w.lastUpdated
	to.Watching = w.watching
	return to
}
