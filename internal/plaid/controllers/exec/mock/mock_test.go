package mock

import (
	"context"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/controllers/exec"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/controllers/logdrain"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestMockInvocation(t *testing.T) {
	t.Run("Given an a Mock Invocation controller integrated into the system", func(t *testing.T) {
		testContext, done := context.WithTimeout(context.Background(), 2*time.Second)
		t.Cleanup(done)

		plaid := withPlaidTest(t, testContext)
		config, tree := logdrain.NewLogDrainSystem(plaid.storeController)
		plaid.testing.AttachController("logging", tree)
		mockExec := AttachMockInvocationController(testContext, plaid.testing, config)

		t.Run("When a new process resource is created", func(t *testing.T) {
			ref := resources.Meta{
				Type: exec.InvocationAlphaV1Type,
				Name: faker.Word(),
			}
			require.NoError(t, plaid.store.Create(testContext, ref, &exec.InvocationAlphaV1Spec{
				Exec: "echo test invocation",
			}))

			t.Run("Then it should not have been started", func(t *testing.T) {
				var status exec.InvocationAlphaV1Status
				require.NoError(t, WaitOnStatus(testContext, plaid.store, ref, &status, func() bool {
					return true
				}))

				assert.Nil(t, status.Started, "should not have been started")
			})

			t.Run("And the process is requested", func(t *testing.T) {
				mock, has := mockExec.For(testContext, ref)
				require.True(t, has)
				require.NotNil(t, mock)

				t.Run("And is started", func(t *testing.T) {
					assert.NoError(t, mock.Start(testContext))

					t.Run("Then the status is updated", func(t *testing.T) {
						var status exec.InvocationAlphaV1Status
						require.NoError(t, WaitOnStatus(testContext, plaid.store, ref, &status, func() bool {
							return status.Started != nil
						}))

						assert.NotNil(t, status.Started, "started time should not be nil")
					})

					t.Run("And a mock line is emitted from stdout", func(t *testing.T) {
						words := faker.Sentence()

						require.NoError(t, mock.Write(testContext, logdrain.LogEntry{
							When:    time.Now(),
							Message: words,
						}))

						t.Run("Then it can be retrieved from the logs", func(t *testing.T) {

						})
					})
				})
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

type Waiter struct {
	MaximumWait time.Duration
	Interval    time.Duration
	Check       func(ctx context.Context) (bool, error)
}

func (w *Waiter) WaitOn(parent context.Context) (bool, error) {
	timeBound, done := context.WithTimeout(parent, w.MaximumWait)
	defer done()

	clock := time.NewTimer(w.Interval)
	for {
		completed, err := w.Check(timeBound)
		if err != nil {
			return completed, err
		}
		if completed {
			return true, nil
		}
		select {
		case <-timeBound.Done():
			return false, timeBound.Err()
		case <-clock.C:
		}
	}
}

func WaitOnStatus(parent context.Context, client *resources.Client, ref resources.Meta, status any, predicate func() bool) error {
	w := &Waiter{
		MaximumWait: 100 * time.Millisecond,
		Interval:    5 * time.Millisecond,
		Check: func(ctx context.Context) (bool, error) {
			exists, err := client.GetStatus(ctx, ref, status)
			if err != nil {
				return false, err
			}
			if !exists {
				return false, nil
			}
			return predicate(), nil
		},
	}
	_, err := w.WaitOn(parent)
	return err
}
