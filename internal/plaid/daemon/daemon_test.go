package daemon

import (
	"context"
	"encoding/json"
	"github.com/meschbach/plaid/internal/plaid/daemon/wire"
	"github.com/meschbach/plaid/internal/plaid/resources"
	"github.com/go-faker/faker/v4"
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
	wire.RegisterResourceControllerServer(s, &ResourceService{
		client: plaid.Controller.Client(),
	})
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

	client := wire.NewResourceControllerClient(conn)
	e := exampleEntity{Words: faker.Word()}
	payload, err := json.Marshal(e)
	require.NoError(t, err)

	ref := &wire.Meta{
		Kind: &wire.Type{
			Kind:    faker.URL(),
			Version: faker.Date(),
		},
		Name: faker.Word(),
	}
	_, err = client.Create(ctx, &wire.CreateResourceIn{
		Target: ref,
		Spec:   payload,
	})
	require.NoError(t, err)
	out, err := client.Get(ctx, &wire.GetIn{Target: ref})
	require.NoError(t, err)
	if assert.True(t, out.Exists, "exists") {
		assert.Equal(t, payload, out.Spec)
	}
}

type exampleEntity struct {
	Words string `json:"words"`
}
