package local

import (
	"context"
	"github.com/meschbach/plaid/internal/junk"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/internal/plaid/controllers/logdrain"
	"github.com/meschbach/plaid/internal/plaid/resources"
	"github.com/go-faker/faker/v4"
	"github.com/meschbach/go-junk-bucket/pkg/reactors"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thejerf/suture/v4"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"testing"
	"time"
)

func TestLocalInvocationController(t *testing.T) {
	t.Run("Given an a Local Invocation controller integrated into the system", func(t *testing.T) {
		onCleanupTracing := junk.SetupTestTracing(t)
		t.Cleanup(func() {
			shutdownCtx, done := context.WithTimeout(context.Background(), 1*time.Second)
			defer done()
			onCleanupTracing(shutdownCtx)
		})
		//timedTestContext, done := context.WithTimeout(context.Background(), 3*time.Second)
		//t.Cleanup(done)
		timedTestContext := context.Background()
		testContext, testSpan := tracer.Start(timedTestContext, t.Name())
		t.Cleanup(func() {
			testSpan.End()
		})

		plaid := withPlaidTest(t, testContext)
		loggerConfig, tree := logdrain.NewLogDrainSystem(plaid.storeController)
		plaid.testing.AttachController("logging", tree)

		procTree := suture.NewSimple("local-exec-procs")
		controllerInstance := &Controller{
			storage:         plaid.storeController,
			logging:         loggerConfig,
			procSupervisors: procTree,
		}
		plaid.testing.AttachController("local-exec", controllerInstance)
		plaid.testing.AttachController("local-exec-procs", procTree)

		t.Run("When a new process resource is created", func(t *testing.T) {
			testContext, _ = traceSubtest(t, testContext)
			ref := resources.Meta{
				Type: exec.InvocationAlphaV1Type,
				Name: faker.Word(),
			}

			ticker := &reactors.Ticked[string]{}
			loggingClientContext := reactors.WithReactor[string](testContext, ticker)
			logs := logdrain.BuildClient[string](testContext, loggerConfig, ticker)
			outputDrain := streams.NewBuffer[logdrain.LogEntry](32, streams.WithBufferTracePrefix[logdrain.LogEntry]("test.stdout"))
			logs.RegisterDrain(loggingClientContext, outputDrain)
			more, err := ticker.Tick(testContext, 32, "")
			assert.False(t, more, "should have consumed all events")
			assert.NoError(t, err)

			require.NoError(t, plaid.store.Create(testContext, ref, &exec.InvocationAlphaV1Spec{
				Exec: "echo test invocation",
			}))

			t.Run("Then the process finishes eventually", func(t *testing.T) {
				testContext, _ = traceSubtest(t, testContext)
				var status exec.InvocationAlphaV1Status
				completed, err := resources.WaitOn(testContext, resources.ForStatusState(plaid.store, ref, &status, func(status *exec.InvocationAlphaV1Status) (bool, error) {
					return status.Started != nil && status.Finished != nil && status.ExitStatus != nil, nil
				}))
				require.NoError(t, err)
				if !assert.True(t, completed, "shell command executed") {
					return
				}

				assert.NotNil(t, status.Started, "should have started")
				assert.NotNil(t, status.Finished, "should have completed")
				assert.NotNil(t, status.ExitStatus, "should have exited cleanly")
			})

			t.Run("Then the output should have been as expected", func(t *testing.T) {
				t.Skip("Some sort of hidden timing problem")
				testContext, _ = traceSubtest(t, testContext)
				output := make([]logdrain.LogEntry, 32)
				for {
					_, err := ticker.Tick(testContext, 32, "")
					require.NoError(t, err)

					count, err := outputDrain.ReadSlice(testContext, output)
					if count > 0 {
						output = output[0:count]
						break
					}
					assert.ErrorIs(t, err, streams.UnderRun)
					select {
					case <-testContext.Done():
						require.NoError(t, testContext.Err())
					default:
						time.Sleep(100 * time.Millisecond)
					}
				}
				//todo: revisit logging system
				if assert.Len(t, output, 1, "has stdout messages") {
					assert.Equal(t, "test invocation", output[0].Message, "has expected output")
				}
			})
		})
	})
}

type plaidHarness struct {
	storeController *resources.Controller
	store           *resources.Client
	testing         *resources.TestSubsystem
}

func withPlaidTest(t *testing.T, ctx context.Context) *plaidHarness {
	core := resources.WithTestSubsystem(t, ctx)
	store := core.Store
	storeController := core.Controller

	return &plaidHarness{
		storeController: storeController,
		store:           store,
		testing:         core,
	}
}

func traceSubtest(t *testing.T, parentContext context.Context) (context.Context, trace.Span) {
	ctx, span := tracer.Start(parentContext, t.Name())
	t.Cleanup(func() {
		if t.Failed() {
			span.SetStatus(codes.Error, "test failed")
		}
		span.End()
	})
	return ctx, span
}
