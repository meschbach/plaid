package resources

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestResourceSystem(t *testing.T) {
	t.Run("Can create and recall resource", func(t *testing.T) {
		//todo: extract deadline from test
		ctx, done := context.WithTimeout(context.Background(), 1*time.Second)
		defer done()

		res := WithTestSubsystem(t, ctx)
		t.Cleanup(res.SystemDone)

		controller := res.Controller

		exampleMeta := Meta{
			Type: Type{
				Kind:    "test",
				Version: "alpha/v1",
			},
			Name: "test",
		}
		exampleBody := []byte("test")
		client := controller.Client()
		require.NoError(t, client.CreateBytes(ctx, exampleMeta, exampleBody))

		body, has, err := client.GetBytes(ctx, exampleMeta)
		require.NoError(t, err)
		assert.Truef(t, has, "has node")
		assert.Equal(t, exampleBody, body)
	})

	t.Run("ResourceManager can watch changes", func(t *testing.T) {
		exampleMeta := Meta{
			Type: Type{
				Kind:    "test",
				Version: "alpha/v1",
			},
			Name: "test",
		}

		//todo: extract deadline from test
		ctx, done := context.WithTimeout(context.Background(), 1*time.Second)
		defer done()
		t.Cleanup(done)

		res := WithTestSubsystem(t, ctx)

		controller := res.Controller

		startSync := make(chan interface{})
		resourceManager := &testResourceManager{
			watchType:   exampleMeta.Type,
			controller:  controller,
			startedSync: startSync,
			allSeen:     make(chan chan []ResourceChanged, 1),
		}
		res.AttachController("test-resource-manager", resourceManager)

		select {
		case <-startSync:
		case <-ctx.Done():
			require.NoError(t, ctx.Err())
			require.Fail(t, "timed out waiting for test harness actor to start")
		}

		exampleBody := []byte("test")
		client := controller.Client()
		require.NoError(t, client.CreateBytes(ctx, exampleMeta, exampleBody))
		sync := make(chan []ResourceChanged, 1)
		resourceManager.allSeen <- sync
		seenChanges := <-sync

		if assert.Len(t, seenChanges, 1) {
			assert.Equal(t, CreatedEvent, seenChanges[0].Operation)
			assert.Equal(t, "test", seenChanges[0].Which.Name)
		}
	})
}

type testResourceManager struct {
	watchType   Type
	seenChanges []ResourceChanged
	controller  *Controller
	startedSync chan interface{}
	allSeen     chan chan []ResourceChanged
}

func (t *testResourceManager) Serve(ctx context.Context) error {
	client := t.controller.Client()
	watch, err := client.Watch(ctx, t.watchType)
	close(t.startedSync)
	if err != nil {
		return err
	}

	for {
		select {
		case sync := <-t.allSeen:
			sync <- t.seenChanges
		case change := <-watch:
			t.seenChanges = append(t.seenChanges, change)
		case <-ctx.Done():
			return nil
		}
	}
}
