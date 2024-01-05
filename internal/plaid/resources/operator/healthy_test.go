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

var MockHealthyType = resources.Type{
	Kind:    "mock-health.plaid.meschbach.com",
	Version: "1",
}

type HealthySpec struct {
}

func TestHealthyDependencies(t *testing.T) {
	t.Run("Given an empty set to watch", func(t *testing.T) {
		testContext, done := context.WithTimeout(context.Background(), 1*time.Second)
		t.Cleanup(done)
		plaid := withPlaidTest(t, testContext)
		store := plaid.store
		w, err := store.Watcher(testContext)
		require.NoError(t, err)

		h := NewHealthDependencies()

		t.Run("Then the set is healthy", func(t *testing.T) {
			healthy, _, err := h.Reconcile(testContext, store)
			require.NoError(t, err)
			assert.True(t, healthy)
		})

		t.Run("When given a healthy dependency; then it is healthy", func(t *testing.T) {
			dep := resources.Meta{
				Type: MockHealthyType,
				Name: faker.Word(),
			}
			requireCreateWithStatus(t, testContext, store, dep, HealthySpec{}, HealthySignal{Healthy: true})

			deps := []resources.Meta{
				dep,
			}
			healthy, _, err := h.Update(testContext, store, w, deps)
			require.NoError(t, err)
			assert.True(t, healthy)
		})

		t.Run("When given a missing dependency", func(t *testing.T) {
			dep := resources.Meta{
				Type: MockHealthyType,
				Name: faker.Name(),
			}

			deps := []resources.Meta{
				dep,
			}
			healthy, _, err := h.Update(testContext, store, w, deps)
			require.NoError(t, err)
			assert.False(t, healthy, "not all dependencies are healthy")

			t.Run("Then the updated set is not healthy", func(t *testing.T) {
				healthy, _, err := h.Reconcile(testContext, store)
				require.NoError(t, err)
				assert.False(t, healthy, "should be unhealthy")
			})

			t.Run("And the dependency is created unhealthy", func(t *testing.T) {
				requireCreateWithStatus(t, testContext, store, dep, HealthySpec{}, HealthySignal{Healthy: false})

				healthy, _, err := h.Reconcile(testContext, store)
				require.NoError(t, err)
				assert.False(t, healthy, "should be unhealthy")

				t.Run("When it becomes healthy", func(t *testing.T) {
					requireUpdateStatus(t, testContext, store, dep, HealthySignal{Healthy: true})

					t.Run("Then is notified of change", func(t *testing.T) {
						hasChange := false
						onChange := func(ctx context.Context) error {
							hasChange = true
							return err
						}
						h.OnChange = onChange
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

					t.Run("Then reconciliation is healthy", func(t *testing.T) {
						healthy, _, err := h.Reconcile(testContext, store)
						require.NoError(t, err)
						assert.True(t, healthy, "should be unhealthy")
					})
				})
			})
		})
	})
}

func requireUpdateStatus(t *testing.T, ctx context.Context, store *resources.Client, ref resources.Meta, status any) {
	t.Helper()
	exists, err := store.UpdateStatus(ctx, ref, status)
	require.NoError(t, err)
	require.True(t, exists)
}

func requireCreateWithStatus(t *testing.T, ctx context.Context, store *resources.Client, ref resources.Meta, spec any, status any) {
	t.Helper()
	require.NoError(t, store.Create(ctx, ref, spec))
	requireUpdateStatus(t, ctx, store, ref, status)
}
