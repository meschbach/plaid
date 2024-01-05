package resources

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/thejerf/suture/v4"
	"testing"
)

type TestSubsystem struct {
	Controller *Controller
	Store      *Client
	SystemDone func()
	Logger     chan<- string
	treeRoot   *suture.Supervisor
}

func (t *TestSubsystem) AttachController(name string, s suture.Service) {
	watch := suture.NewSimple(name)
	watch.Add(s)

	t.treeRoot.Add(watch)
}

func WithTestSubsystem(t *testing.T, parentContext context.Context) *TestSubsystem {
	supervisionTreeContext, onSupervisionTreeDone := context.WithCancel(parentContext)
	t.Cleanup(onSupervisionTreeDone)
	//Logger
	msg := make(chan string, 64)
	go func() {
		defer func() {
			recover()
		}()
		for {
			select {
			case m := <-msg:
				t.Logf("%s", m)
			case <-parentContext.Done():
				return
			}
		}
	}()

	//core resource system
	storeController := NewController()
	resSupervisor := suture.NewSimple("resources")
	resSupervisor.Add(storeController)
	store := storeController.Client()

	//build primary supervisor tree
	root := suture.NewSimple(t.Name())
	root.Add(resSupervisor)
	systemDone := root.ServeBackground(supervisionTreeContext)

	//return
	return &TestSubsystem{
		Controller: storeController,
		Store:      store,
		SystemDone: func() {
			t.Helper()
			require.ErrorIs(t, <-systemDone, context.Canceled)
		},
		Logger:   msg,
		treeRoot: root,
	}
}
