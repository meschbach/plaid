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

type exampleEntity struct {
	Words string `json:"words"`
}

type exampleStatus struct {
	Response string `json:"response"`
}

func TestSystem(t *testing.T) {
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

	clientSide.Run("Watch type", func(t *testing.T, clientSide *optest.System, ctx context.Context) {
		exampleKind := resources.FakeType()
		observer := clientSide.ObserveType(ctx, exampleKind)

		ref := resources.FakeMetaOf(exampleKind)
		entityStatus := exampleEntity{Words: "z"}
		clientSide.Run("When creating a new resource", func(t *testing.T, clientSide *optest.System, ctx context.Context) {
			anyEvent := observer.AnyEvent.Fork()
			create := observer.Create.Fork()
			clientSide.MustCreate(ctx, ref, entityStatus)

			anyEvent.Wait(ctx)
			create.Wait(ctx)
		})

		clientSide.Run("When updating status of an exiting resource", func(t *testing.T, clientSide *optest.System, ctx context.Context) {
			status := exampleStatus{Response: "destroyed systems"}
			anyChange := observer.AnyEvent.Fork()
			statusChange := observer.UpdateStatus.Fork()
			clientSide.MustUpdateStatus(ctx, ref, status)

			anyChange.Wait(ctx)
			statusChange.Wait(ctx)
		})

		clientSide.Run("When the resource is deleted", func(t *testing.T, clientSide *optest.System, ctx context.Context) {
			deleteOp := observer.Delete.Fork()
			serviceSide.MustDelete(ctx, ref)
			deleteOp.Wait(ctx)
		})
	})
}
