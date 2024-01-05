package mock

import (
	"context"
	"errors"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/internal/plaid/controllers/logdrain"
	"github.com/meschbach/plaid/internal/plaid/resources"
	"github.com/meschbach/plaid/internal/plaid/resources/operator"
	"github.com/meschbach/go-junk-bucket/pkg/emitter"
	"github.com/meschbach/go-junk-bucket/pkg/reactors"
	"github.com/meschbach/go-junk-bucket/pkg/reactors/futures"
	"go.opentelemetry.io/otel/attribute"
	"time"
)

// ExecEngine provides an execution engine to be controlled via tests to verify expected changes within a system.
type ExecEngine struct {
	events    *emitter.Dispatcher[*engineState]
	c         *resources.Controller
	reactor   *reactors.Channel[*engineState]
	reactorIn <-chan reactors.ChannelEvent[*engineState]
	logging   *logdrain.Client[*engineState]
}

func (m *ExecEngine) Serve(ctx context.Context) error {
	store := m.c.Client()
	state := &engineState{
		controller: store,
		procs:      resources.NewMetaContainer[Proc](),
	}

	//todo: used for anything?
	watcher, err := store.Watcher(ctx)
	if err != nil {
		return err
	}

	bridge := operator.NewKindBridge[exec.InvocationAlphaV1Spec, exec.InvocationAlphaV1Status, Proc](exec.InvocationAlphaV1Type, &alpha1Ops{
		push:   store,
		engine: m,
	})
	av1Event, err := bridge.Setup(ctx, store)

	for {
		select {
		case e := <-av1Event:
			if err := bridge.Dispatch(ctx, store, e); err != nil {
				return err
			}
		case e := <-m.reactorIn:
			if err := m.reactor.Tick(ctx, e, state); err != nil {
				return err
			}
		case e := <-watcher.Feed:
			if err := watcher.Digest(ctx, e); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (m *ExecEngine) forResource(ctx context.Context, ref resources.Meta) (*Proc, bool) {
	type result struct {
		value *Proc
		has   bool
	}

	doneSignal := make(chan interface{}, 1)
	out := &Promsie[result]{
		onChange: &emitter.Dispatcher[*Promsie[result]]{},
	}

	p := futures.PromiseFuncOn[*engineState, result](ctx, m.reactor, func(ctx context.Context, state *engineState) (result, error) {
		ptr, has := state.procs.Find(ref)
		if !has {
			var sub *emitter.Subscription[*engineState]
			sub = m.events.OnceE(func(ctx context.Context, event *engineState) error {
				ptr, has := state.procs.Find(ref)
				if has {
					m.events.Off(sub)
					return out.Resolve(ctx, result{ptr, has})
				}
				return nil
			})
		}
		return result{
			value: ptr,
			has:   has,
		}, nil
	})
	r, err := p.Await(ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, false
		}
		panic(err)
	}
	if r.Result.has {
		return r.Result.value, r.Result.has
	}
	_, err = out.Then(ctx, func(ctx context.Context, event result) error {
		close(doneSignal)
		return nil
	})
	if err != nil {
		panic(err)
	}
	select {
	case <-doneSignal:
	case <-ctx.Done():
		return nil, false
	}
	return out.result.value, out.result.has
}

func (m *ExecEngine) FinishAll(parent context.Context, storage *resources.Client) (int, error) {
	ctx, span := tracer.Start(parent, "mockExecEngine.FinishAll")
	defer span.End()
	allInvocations, err := storage.List(ctx, exec.InvocationAlphaV1Type)
	if err != nil {
		return 0, err
	}
	span.SetAttributes(attribute.Int("found.count", len(allInvocations)))

	updates := 0
	for _, invocation := range allInvocations {
		updated := false
		proc, _ := m.For(ctx, invocation)
		//no proc yet
		if proc == nil {
			continue
		}
		if proc.startedAt == nil {
			span.AddEvent("starting " + invocation.String())
			if err := proc.Start(ctx); err != nil {
				return updates, err
			}
			updated = true
		}

		if proc.finishedAt == nil {
			span.AddEvent("finishing " + invocation.String())
			if err := proc.Finish(ctx); err != nil {
				return updates, err
			}
			updated = true
		}

		if updated {
			updates++
		}
	}
	span.SetAttributes(attribute.Int("change.count", updates))
	return updates, nil
}

// For locates the given resource
func (m *ExecEngine) For(ctx context.Context, ref resources.Meta) (*Proc, bool) {
	return m.forResource(ctx, ref)
}

func (m *ExecEngine) WaitFor(ctx context.Context, ref resources.Meta) (*Proc, bool) {
	for {
		if p, has := m.For(ctx, ref); has {
			return p, true
		}

		select {
		case <-ctx.Done():
			panic(ctx.Err())
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}
