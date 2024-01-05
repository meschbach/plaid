package logdrain

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/resources"
	"github.com/meschbach/go-junk-bucket/pkg/emitter"
	"github.com/meschbach/go-junk-bucket/pkg/reactors"
	"github.com/meschbach/go-junk-bucket/pkg/reactors/futures"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
)

type Sources interface {
	GetLogStream(name string) (streams.Source[string], error)
}

type registeredSource struct {
	source Sources
}

type Registry struct {
	feeds    resources.MetaContainer[registeredSource]
	onChange emitter.Dispatcher[resources.Meta]
	ticks    <-chan reactors.ChannelEvent[*Registry]
	ticker   *reactors.Channel[*Registry]
}

func (r *Registry) Register(ctx context.Context, ref resources.Meta, from Sources) {
	r.ticker.ScheduleStateFunc(ctx, func(ctx context.Context, r *Registry) error {
		//fmt.Printf("[log-drain/registry] Registering %s\n", ref)
		r.feeds.Upsert(ref, &registeredSource{source: from})
		return r.onChange.Emit(ctx, ref)
	})
}

func (r *Registry) Unregister(ctx context.Context, ref resources.Meta) {
	r.ticker.ScheduleStateFunc(ctx, func(ctx context.Context, r *Registry) error {
		r.feeds.Delete(ref)
		return r.onChange.Emit(ctx, ref)
	})
}

func (r *Registry) Locate(ctx context.Context, ref resources.Meta) (Sources, bool) {
	type output struct {
		from Sources
		has  bool
	}
	p := futures.PromiseFuncOn[*Registry, output](ctx, r.ticker, func(ctx context.Context, state *Registry) (output, error) {
		out := output{}
		registration, has := r.feeds.Find(ref)
		if has {
			out.has = true
			out.from = registration.source
		} else {
			out.has = false
		}
		return out, nil
	})
	result, err := p.Await(ctx) //todo: use newer primitives
	if err != nil {
		panic(err)
	}
	return result.Result.from, result.Result.has
}

func (r *Registry) Observe(ctx context.Context, ref resources.Meta, onChange func(sources Sources, has bool) error) error {
	p := futures.PromiseFuncOn[*Registry, error](ctx, r.ticker, func(ctx context.Context, state *Registry) (error, error) {
		//fmt.Printf("[log-drain/registry] Observing %s\n", ref)
		publish := func() error {
			registration, has := state.feeds.Find(ref)
			//fmt.Printf("[log-drain/registry] Publishing change to %s has? %t  -- %#v\n", ref, has, registration)
			var source Sources
			if registration != nil {
				source = registration.source
			}
			return onChange(source, has)
		}
		r.onChange.OnE(func(ctx context.Context, event resources.Meta) error {
			if event == ref {
				return publish()
			} else {
				return nil
			}
		})
		if err := publish(); err != nil {
			return err, nil
		}
		return nil, nil
	})
	result, err := p.Await(ctx)
	if err != nil {
		return err
	}
	return result.Result
}

func (r *Registry) Serve(ctx context.Context) error {
	for {
		select {
		case e := <-r.ticks:
			if err := r.ticker.Tick(ctx, e, r); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
