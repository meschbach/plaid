package fileWatch

import (
	"github.com/meschbach/plaid/resources"
	"github.com/thejerf/suture/v4"
)

func NewFileWatch(parent *suture.Supervisor, res *resources.Controller) {
	notify := newLocalFSWatcher()
	fsNotifySupervisor := suture.NewSimple("file-watcher-fsnotify")
	fsNotifySupervisor.Add(notify)

	fileWatchSupervisor := suture.NewSimple("file-watcher-resources")
	fileWatchSupervisor.Add(NewController(res, notify))

	parent.Add(fileWatchSupervisor)
	parent.Add(fsNotifySupervisor)
}
