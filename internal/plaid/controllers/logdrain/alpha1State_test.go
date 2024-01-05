package logdrain

import (
	"context"
	"git.meschbach.com/mee/platform.git/plaid/internal/junk"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources/operator"
	"github.com/go-faker/faker/v4"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"testing"
	"time"
)

type mockKindBridgeState struct {
	resourceChangedCalls []resources.Meta
}

func (m *mockKindBridgeState) OnResourceChange(ctx context.Context, which resources.Meta) error {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("mockKindBridgeState.OnResourceChange", trace.WithAttributes(attribute.Stringer("which", which)))

	m.resourceChangedCalls = append(m.resourceChangedCalls, which)
	return nil
}

func TestAlpha1State(t *testing.T) {
	t.Run("Given an initial state", func(t *testing.T) {
		rootCtx, done := context.WithTimeout(context.Background(), 2*time.Second)
		defer done()
		shutdownTracing := junk.SetupTestTracing(t)
		defer func() {
			shutdown, done := context.WithTimeout(context.Background(), 10*time.Second)
			defer done()
			shutdownTracing(shutdown)
		}()

		ctx, span := tracing.Start(rootCtx, "TestAlpha1State")
		defer span.End()

		mockBridge := &mockKindBridgeState{}
		bridge := &operator.KindBridgeState{OnResourceChange: func(ctx context.Context, which resources.Meta) error {
			return mockBridge.OnResourceChange(ctx, which)
		}}
		state := &alpha1State{
			bridge: bridge,
			self:   resources.Meta{},
		}

		t.Run("Then the state is not connected", func(t *testing.T) {
			status := state.toStatus()
			assert.Equal(t, Unconnected, status.Pipe)
		})

		t.Run("When connecting an empty source then empty sink", func(t *testing.T) {
			input := streams.NewBuffer[LogEntry](12)
			output := streams.NewBuffer[LogEntry](12)

			require.NoError(t, state.updateSource(ctx, input))
			require.NoError(t, state.updateDrain(ctx, output))

			t.Run("Then it is connected", func(t *testing.T) {
				status := state.toStatus()
				assert.Equal(t, Connected, status.Pipe)
			})

			t.Run("Then the output stream produces a single output", func(t *testing.T) {
				entry := LogEntry{
					When:    time.Now(),
					Message: faker.Sentence(),
				}
				require.NoError(t, input.Write(ctx, entry))
				assert.Equal(t, entry, output.Output[0])
			})
			t.Run("Then an update is scheduled once", func(t *testing.T) {
				if assert.Len(t, mockBridge.resourceChangedCalls, 1) {
				}
			})
		})
	})
}
