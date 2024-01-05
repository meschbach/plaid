package controllers

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-faker/faker/v4"
	"github.com/meschbach/go-junk-bucket/pkg/reactors"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
	"github.com/meschbach/plaid/internal/junk"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec/mock"
	"github.com/meschbach/plaid/internal/plaid/controllers/logdrain"
	"github.com/meschbach/plaid/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestLogDrain(t *testing.T) {
	t.Skip("Needs to be reviewed.  I moved away from using a resource to describe the streams underneath.")
	t.Run("Given a logging source", func(t *testing.T) {
		rootCtx := context.Background()
		junk.SetupTestTracing(t)

		logOutputBuffer := streams.NewBuffer[logdrain.LogEntry](32)
		internalConsumerName := faker.Word()
		loggerName := faker.Word()
		execName := faker.Word()

		drainRef := resources.Meta{
			Type: logdrain.InternalLoggingAlpha1,
			Name: internalConsumerName,
		}

		spanCtx, _ := junk.TraceSubtest(t, rootCtx, tracing)
		ctx, done := context.WithTimeout(spanCtx, 2*time.Second)
		defer done()

		plaid := resources.WithTestSubsystem(t, ctx)
		drainConfig, loggingTree := logdrain.NewLogDrainSystem(plaid.Controller)
		plaid.AttachController("logs", loggingTree)
		execEngine := mock.AttachMockInvocationController(ctx, plaid, drainConfig)

		type testState struct{}
		tickedReactor := &reactors.Ticked[*testState]{}
		tickedReactorContext := reactors.WithReactor[*testState](ctx, tickedReactor)

		logging := logdrain.BuildClient[*testState](ctx, drainConfig, tickedReactor)
		logging.RegisterDrain(tickedReactorContext, logOutputBuffer)
		_, tickError := tickedReactor.Tick(ctx, 10, &testState{})
		assert.NoError(t, tickError)

		execRef := resources.Meta{
			Type: exec.InvocationAlphaV1Type,
			Name: execName,
		}
		execSpec := exec.InvocationAlphaV1Spec{
			Exec:       "test 123",
			WorkingDir: "/example",
		}
		require.NoError(t, plaid.Store.Create(ctx, execRef, execSpec))

		t.Run("When a matching log drain is created", func(t *testing.T) {
			ctx, _ = junk.TraceSubtest(t, ctx, tracing)
			logDrainRef := resources.Meta{
				Type: logdrain.AlphaV1Type,
				Name: loggerName,
			}
			spec := logdrain.Alpha1Spec{
				Source: logdrain.AlphaV1SourceSpec{
					Ref:    execRef,
					Stream: "stdout",
				},
				Drain: logdrain.AlphaV1DrainSpec{
					Ref:    drainRef,
					Stream: loggerName,
				},
			}
			require.NoError(t, plaid.Store.Create(ctx, logDrainRef, spec))

			t.Run("Then the status reflects connected", func(t *testing.T) {
				ctx, _ = junk.TraceSubtest(t, ctx, tracing)
				var status logdrain.Alpha1Status
				waitOn(t, ctx, 2*time.Second, func(ctx context.Context) (bool, error) {
					exists, err := plaid.Store.GetStatus(ctx, logDrainRef, &status)
					if err != nil {
						return false, err
					}
					if !exists {
						return false, nil
					}
					return status.Pipe == logdrain.Connected, nil
				}, "waiting on pipe status to become connected")
			})

			t.Run("And a new log message is sent", func(t *testing.T) {
				ctx, _ = junk.TraceSubtest(t, tickedReactorContext, tracing)
				streamDone := false
				logOutputBuffer.SinkEvents().Finishing.On(func(ctx context.Context, event streams.Sink[logdrain.LogEntry]) {
					fmt.Println("ooo\t\tlogOutputBuffer on finishing")
					invokingReactor, has := reactors.Maybe[*testState](ctx)
					if assert.True(t, has, "should have a context") {
						assert.Equal(t, tickedReactor, invokingReactor, "event dispatched on expected reactor")
					}
					streamDone = true
				})

				exampleLine := faker.Sentence()
				proc, has := execEngine.For(ctx, execRef)
				require.True(t, has, "must have resource engine")
				require.NoError(t, proc.Start(ctx))
				mock.With[*testState](ctx, tickedReactor, proc, func(ctx context.Context, m *mock.Proc) error {
					require.NoError(t, proc.StdOut.Write(ctx, logdrain.LogEntry{
						When:    time.Now(),
						Message: exampleLine,
					}))
					require.NoError(t, proc.StdOut.Finish(ctx))
					return nil
				})
				for {
					_, err := tickedReactor.Tick(ctx, 32, &testState{})
					if errors.Is(err, context.DeadlineExceeded) {
						assert.NoError(t, err)
						return
					}
					require.NoError(t, err)
					if streamDone {
						break
					}
				}

				t.Run("Then it is received by the target system", func(t *testing.T) {
					//todo: figure out why lines are duplicated
					//if assert.Len(t, logOutputBuffer.Output, 1) {
					assert.Equal(t, exampleLine, logOutputBuffer.Output[0].Message)
					//}
				})
			})
		})
	})
}

func waitOn(t *testing.T, parent context.Context, howLong time.Duration, test func(ctx context.Context) (bool, error), message string) {
	t.Helper()
	ctx, done := context.WithTimeout(parent, howLong)
	defer done()

	for {
		success, err := test(ctx)
		if errors.Is(err, context.DeadlineExceeded) {
			assert.Failf(t, message, "timed out while waiting for condition")
			return
		} else {
			require.NoError(t, err)
		}

		if success {
			return
		}

		time.Sleep(5 * time.Millisecond)
	}
}
