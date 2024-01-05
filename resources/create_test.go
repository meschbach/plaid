package resources

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCreateOperations(t *testing.T) {
	t.Run("Given a claimer and annotation", func(t *testing.T) {
		baseCtx, done := context.WithTimeout(context.Background(), 1*time.Second)
		t.Cleanup(done)
		plaid := WithTestSubsystem(t, baseCtx)

		otherRef := FakeMeta()
		annotation := map[string]string{
			"test": "kernel mode",
		}
		fake := FakeMeta()
		require.NoError(t, plaid.Store.Create(baseCtx, fake, ExampleResource{Enabled: true}, ClaimedBy(otherRef), WithAnnotations(annotation)))

		t.Run("When queried for the resource metadata", func(t *testing.T) {
			metadata, exists, err := plaid.Store.GetMetadataFor(baseCtx, fake)
			require.NoError(t, err)
			if assert.True(t, exists) && assert.NotNil(t, metadata.Annotations) {
				value, exists := metadata.Annotations["test"]
				if assert.True(t, exists) {
					assert.Equal(t, "kernel mode", value)
				}
			}
		})
	})
}
