package sys

import (
	"context"
	"github.com/meschbach/plaid/ipc/grpc/reswire"
	"github.com/meschbach/plaid/ipc/grpc/reswire/client"
	"github.com/meschbach/plaid/ipc/grpc/reswire/service"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/optest"
	"github.com/stretchr/testify/require"
	"github.com/thejerf/suture/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"testing"
)

func TestWatcher(t *testing.T) {
	ctx, serviceSide := optest.New(t)

	s := grpc.NewServer()
	reswire.RegisterResourceControllerServer(s, service.New(serviceSide.Legacy.Controller.Client()))

	//setup server
	buffer := 1024 * 1024
	listener := bufconn.Listen(buffer)
	//todo: should probably be wrapped properly in a service
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
	clientSideTree := suture.NewSimple("wire.storage")
	serviceSide.Legacy.AttachController("wire.storage", clientSideTree)
	clientSystem := client.New(conn, clientSideTree)
	clientSide := optest.From(t, ctx, clientSystem)

	t.Run("Watch type", func(t *testing.T) {
		exampleKind := resources.FakeType()
		observer := clientSide.ObserveType(ctx, exampleKind)

		ref := resources.FakeMetaOf(exampleKind)
		t.Run("When creating a new resource", func(t *testing.T) {
			create := observer.Create.Fork()
			clientSide.MustCreate(ctx, ref, exampleEntity{Words: "z"})

			create.Wait(ctx)
		})

		t.Run("When the resource is deleted", func(t *testing.T) {
			deleteOp := observer.Delete.Fork()
			serviceSide.MustDelete(ctx, ref)
			deleteOp.Wait(ctx)
		})
	})
}