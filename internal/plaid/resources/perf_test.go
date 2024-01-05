package resources

import (
	"context"
	"fmt"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
	"github.com/thejerf/suture/v4"
	"testing"
)

var BenchmarkType = Type{
	Kind:    "benchmark.plaid.meschbach.com",
	Version: "alpha.1",
}

type BenchmarkResource struct {
	Test string
}

func BenchmarkClient_Create(b *testing.B) {
	testContext, done := context.WithCancel(context.Background())
	defer done()
	ctx := WithFastCreate(testContext)
	platform := WithPerformanceSystem(b, ctx)
	for i := 0; i < b.N; i++ {
		name := fmt.Sprintf("%s-%d", faker.Name(), i)
		require.NoError(b, platform.Store.Create(ctx, Meta{
			Type: BenchmarkType,
			Name: name,
		}, BenchmarkResource{Test: name}))
	}
}

func WithPerformanceSystem(t *testing.B, ctx context.Context) *TestSubsystem {
	//Logger
	msg := make(chan string, 64)
	go func() {
		<-msg
	}()

	//core resource system
	storeController := NewController()
	resSupervisor := suture.NewSimple("resources")
	resSupervisor.Add(storeController)
	store := storeController.Client()

	//build primary supervisor tree
	root := suture.NewSimple(t.Name())
	root.Add(resSupervisor)
	systemDone := root.ServeBackground(ctx)

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
