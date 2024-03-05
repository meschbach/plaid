package sys

import (
	"context"
	"encoding/json"
	"github.com/go-faker/faker/v4"
	"github.com/meschbach/plaid/ipc/grpc/reswire"
	"github.com/meschbach/plaid/ipc/grpc/reswire/service"
	"github.com/meschbach/plaid/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"testing"
	"time"
)

func TestDaemonSetup(t *testing.T) {
	ctx, done := context.WithTimeout(context.Background(), 1*time.Second)
	defer done()

	plaid := resources.WithTestSubsystem(t, ctx)

	//setup server
	buffer := 1024 * 1024
	listener := bufconn.Listen(buffer)

	s := grpc.NewServer()
	reswire.RegisterResourceControllerServer(s, service.New(plaid.Controller.Client()))
	go func() {
		require.NoError(t, s.Serve(listener))
	}()
	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	conn, _ := grpc.DialContext(ctx, "", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}), grpc.WithInsecure(), grpc.WithBlock())

	client := reswire.NewResourceControllerClient(conn)
	e := exampleEntity{Words: faker.Word()}
	payload, err := json.Marshal(e)
	require.NoError(t, err)

	ref := &reswire.Meta{
		Kind: &reswire.Type{
			Kind:    faker.URL(),
			Version: faker.Date(),
		},
		Name: faker.Word(),
	}
	_, err = client.Create(ctx, &reswire.CreateResourceIn{
		Target: ref,
		Spec:   payload,
	})
	require.NoError(t, err)
	out, err := client.Get(ctx, &reswire.GetIn{Target: ref})
	require.NoError(t, err)
	if assert.True(t, out.Exists, "exists") {
		assert.Equal(t, payload, out.Spec)
	}
}

type exampleEntity struct {
	Words string `json:"words"`
}
