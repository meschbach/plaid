package logger

import (
	"context"
	"fmt"
	"github.com/go-faker/faker/v4"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"sync"
	"testing"
	"time"
)

func TestGRPCLoggerSystemically(t *testing.T) {
	t.Run("Given a connected client and service", func(t *testing.T) {
		scope, testScopeDone := context.WithTimeout(context.Background(), 1*time.Second)
		t.Cleanup(testScopeDone)

		core := &localLogger{
			changes:       sync.Mutex{},
			consumedNames: make(map[string]streams.Sink[bufferedEntry]),
		}
		service := newV1LoggingService(core)

		buffer := 16 * 1024
		listener := bufconn.Listen(buffer)
		grpcServer := grpc.NewServer()
		RegisterV1Server(grpcServer, service)
		go func() {
			require.NoError(t, grpcServer.Serve(listener))
		}()
		t.Cleanup(func() {
			grpcServer.GracefulStop()
		})

		conn, err := grpc.DialContext(scope, "", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}), grpc.WithInsecure(), grpc.WithBlock())
		require.NoError(t, err, "failed to setup the grpc client")

		client := NewV1Client(conn)
		t.Cleanup(func() {
			require.NoError(t, conn.Close())
		})

		t.Run("When exporting a log drain", func(t *testing.T) {
			drainName := faker.Name()
			registrationReply, err := client.RegisterDrain(scope, &RegisterDrainRequest{
				Name: drainName,
			})
			require.NoError(t, err)
			assert.Less(t, int64(0), registrationReply.DrainID, "given a drain id greater than 0")

			t.Run("Then an empty read may occur immediately", func(t *testing.T) {
				items, err := client.ReadDrain(scope, &ReadDrainRequest{
					DrainID: registrationReply.DrainID,
					Offset:  registrationReply.InitialOffset,
					Count:   10,
				})
				require.NoError(t, err)
				assert.Len(t, items.Entries, 0, "no items should be available")
				assert.Equal(t, items.NextOffset, items.BeginningOffset, "offset should not have changed")
			})

			t.Run("When a value is ready for the drain", func(t *testing.T) {
				ctx, ctxDone := context.WithCancel(scope)
				t.Cleanup(ctxDone)
				when := time.Now()
				text := "jingle bells"
				drainNotices, err := client.WatchDrain(ctx)
				require.NoError(t, err)
				require.NoError(t, drainNotices.Send(&WatchDrainRequest{
					DrainID: &registrationReply.DrainID,
				}))
				t.Cleanup(func() {
					require.NoError(t, drainNotices.CloseSend())
				})
				initReply, err := drainNotices.Recv()
				require.NoError(t, err)
				require.Equal(t, int64(0), initReply.Offset, "first offset should always be zero")
				fmt.Println("Setup")

				inputPipe, has := core.findStream(ctx, drainName)
				require.True(t, has, "must have the registered stream")

				require.NoError(t, inputPipe.Write(scope, bufferedEntry{
					when: when,
					text: text,
				}))
				fmt.Println("Written")

				t.Run("Then an event is dispatched", func(t *testing.T) {
					event, err := drainNotices.Recv()
					require.NoError(t, err)
					if assert.NotNil(t, event) {
						assert.Equal(t, int64(1), event.Offset)
					}
				})

				t.Run("Then the value is readable", func(t *testing.T) {
					items, err := client.ReadDrain(scope, &ReadDrainRequest{
						DrainID: registrationReply.DrainID,
						Offset:  registrationReply.InitialOffset,
						Count:   10,
					})
					require.NoError(t, err)
					if assert.Len(t, items.Entries, 1, "no items should be available") {
						item := items.Entries[0]
						assert.Equal(t, when.UTC(), item.When.AsTime())
						assert.Equal(t, text, item.Text)
					}
				})
			})
		})
	})
}
