package mock

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/emitter"
	"sync"
)

// todo: collapse with junk bucket implementation
type Promsie[R any] struct {
	lock     sync.Mutex
	resolved bool
	result   R
	onChange *emitter.Dispatcher[*Promsie[R]]
}

func (p *Promsie[R]) Resolve(ctx context.Context, value R) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.resolved {
		return nil
		//panic("already resolved")
	}

	p.resolved = true
	p.result = value
	return p.onChange.Emit(ctx, p)
}

func (p *Promsie[R]) Then(ctx context.Context, then func(ctx context.Context, result R) error) (*emitter.Subscription[*Promsie[R]], error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.resolved {
		err := then(ctx, p.result)
		return nil, err
	} else {
		out := p.onChange.OnceE(func(ctx context.Context, event *Promsie[R]) error {
			return then(ctx, event.result)
		})
		return out, nil
	}
}
