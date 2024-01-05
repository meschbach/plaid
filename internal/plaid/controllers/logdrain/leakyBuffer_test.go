package logdrain

import (
	"context"
	"github.com/go-faker/faker/v4"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func randInt() int {
	ints, err := faker.RandomInt(-100, 100, 1)
	if err != nil {
		panic(err)
	}
	return ints[0]
}

func TestLeakyBuffer(t *testing.T) {
	t.Run("Given a leaky buffer of size 4", func(t *testing.T) {
		ctx, done := context.WithTimeout(context.Background(), 1*time.Second)
		t.Cleanup(done)

		var sink streams.Sink[int]
		var source streams.Source[LeakyBufferEntry[int]]
		buffer := NewLeakyBuffer[int](4)
		sink = buffer
		source = buffer

		t.Run("When the leaky buffer is given an element", func(t *testing.T) {
			example := randInt()
			require.NoError(t, sink.Write(ctx, example))

			t.Run("Then it is available with no drops", func(t *testing.T) {
				values := make([]LeakyBufferEntry[int], 32)
				count, err := source.ReadSlice(ctx, values)
				if assert.NoError(t, err) {
					assert.Equal(t, 1, count)
					assert.Equal(t, example, values[0].Entry)
					assert.Equal(t, uint(0), values[0].Leaked, "no items should have leaked")
				}
			})
		})
		t.Run("When the buffer is overwhelmed", func(t *testing.T) {
			require.NoError(t, sink.Write(ctx, 0))
			require.NoError(t, sink.Write(ctx, 1))
			require.NoError(t, sink.Write(ctx, 2))
			require.NoError(t, sink.Write(ctx, 3))
			require.NoError(t, sink.Write(ctx, 4))
			require.NoError(t, sink.Write(ctx, 5))

			t.Run("And 4 items are read", func(t *testing.T) {
				values := make([]LeakyBufferEntry[int], 32)
				count, err := source.ReadSlice(ctx, values)
				require.NoError(t, err)
				assert.Equal(t, 4, count, "only 4 items should be available")

				t.Run("Then the first item indicates the drop count", func(t *testing.T) {
					assert.Equal(t, uint(2), values[0].Leaked)
					t.Run("And the correct value", func(t *testing.T) {
						assert.Equal(t, 2, values[0].Entry)
					})
				})

				t.Run("Then all values are as expected", func(t *testing.T) {
					assert.Equal(t, uint(0), values[1].Leaked)
					assert.Equal(t, 3, values[1].Entry)
					assert.Equal(t, uint(0), values[2].Leaked)
					assert.Equal(t, 4, values[2].Entry)
					assert.Equal(t, uint(0), values[3].Leaked)
					assert.Equal(t, 5, values[3].Entry)
				})
			})
		})
	})
}
