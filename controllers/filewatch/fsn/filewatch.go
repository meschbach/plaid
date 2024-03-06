package fsn

import (
	"github.com/meschbach/plaid/controllers/filewatch"
	"github.com/meschbach/plaid/resources"
	"github.com/thejerf/suture/v4"
)

func NewFileWatchSystem(r resources.System) *suture.Supervisor {
	fsn := NewCore()
	controller := filewatch.NewController(r, fsn)
	parent := suture.NewSimple("file-watch")
	parent.Add(controller)
	parent.Add(fsn)
	return parent
}
