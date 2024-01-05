package operator

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/resources"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var MockReadyType = resources.Type{
	Kind:    "mock-ready.plaid.meschbach.com",
	Version: "1",
}

type ReadySpec struct {
}

func TestReadyDependencies(t *testing.T) {
	t.Run("Given an empty set to watch", func(t *testing.T) {
		testContext, done := context.WithTimeout(context.Background(), 1*time.Second)
		t.Cleanup(done)
		plaid := withPlaidTest(t, testContext)
		store := plaid.store
		w, err := store.Watcher(testContext)
		require.NoError(t, err)

		r := NewReadinessObserver()

		t.Run("Then the set is ready", func(t *testing.T) {
			ready, _, err := r.Reconcile(testContext, store)
			require.NoError(t, err)
			assert.True(t, ready)
		})

		t.Run("When given a ready dependency; then it is ready", func(t *testing.T) {
			dep := resources.Meta{
				Type: MockReadyType,
				Name: faker.Word(),
			}
			requireCreateWithStatus(t, testContext, store, dep, ReadySpec{}, ReadySignal{Ready: true})

			deps := []resources.Meta{
				dep,
			}
			ready, err := r.Update(testContext, store, w, deps)
			require.NoError(t, err)
			assert.True(t, ready)
		})

		t.Run("When given a missing dependency", func(t *testing.T) {
			dep := resources.Meta{
				Type: MockReadyType,
				Name: faker.Name(),
			}

			deps := []resources.Meta{
				dep,
			}
			ready, err := r.Update(testContext, store, w, deps)
			require.NoError(t, err)
			assert.False(t, ready, "not all dependencies are ready")

			t.Run("Then the updated set is not ready", func(t *testing.T) {
				ready, results, err := r.Reconcile(testContext, store)
				require.NoError(t, err)
				assert.False(t, ready, "should be not ready, got %#v", results)
			})

			t.Run("And the dependency is created not ready", func(t *testing.T) {
				requireCreateWithStatus(t, testContext, store, dep, ReadySpec{}, ReadySignal{Ready: false})

				ready, readyStatus, err := r.Reconcile(testContext, store)
				require.NoError(t, err)
				assert.False(t, ready, "should be not be ready: %#v", readyStatus)

				t.Run("When it becomes ready", func(t *testing.T) {
					testFeedObservations(t, testContext, w)
					requireUpdateStatus(t, testContext, store, dep, ReadySignal{Ready: true})

					t.Run("Then is notified of change", func(t *testing.T) {
						hasChange := false
						onChange := func(ctx context.Context) error {
							hasChange = true
							return err
						}
						r.OnChange = onChange
						for !hasChange {
							select {
							case e := <-w.Feed:
								require.NoError(t, w.Digest(testContext, e))
							case <-testContext.Done():
								require.NoError(t, testContext.Err())
							}
						}
						assert.True(t, hasChange)
					})

					t.Run("Then reconciliation is ready", func(t *testing.T) {
						ready, _, err := r.Reconcile(testContext, store)
						require.NoError(t, err)
						assert.True(t, ready, "should be ready as dependency is ready")
					})
				})
			})
		})
	})
}

func testFeedObservations(t *testing.T, ctx context.Context, w *resources.ClientWatcher) {
	hasMore := true
	for hasMore {
		select {
		case e := <-w.Feed:
			require.NoError(t, w.Digest(ctx, e))
		case <-ctx.Done():
			require.NoError(t, ctx.Err())
		default:
			hasMore = false
		}
	}
}
