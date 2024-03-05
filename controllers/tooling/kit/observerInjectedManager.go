package kit

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/resources"
	"go.opentelemetry.io/otel/trace"
)

type observerInjectedManager struct {
	target   resources.Meta
	observer resources.Watcher
}

func (p *observerInjectedManager) UpdateStatus(ctx context.Context) error {
	fmt.Printf("Updating %#v\n", p.target)
	event := resources.ResourceChanged{
		Which:     p.target,
		Operation: resources.UpdatedEvent,
		Tracing:   trace.LinkFromContext(ctx),
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case p.observer.Events() <- event:
		fmt.Printf("Dispatched")
		return nil
	}

}
