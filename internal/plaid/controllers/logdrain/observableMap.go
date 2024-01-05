package logdrain

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/emitter"
	"golang.org/x/exp/constraints"
)

type MapKey constraints.Ordered

type MapChangeEvent[K MapKey, V any] struct {
	Source  *ObservableMap[K, V]
	Changed K
}

type ObservableMap[K MapKey, V any] struct {
	OnChange emitter.Dispatcher[MapChangeEvent[K, V]]
	entries  map[K]V
}

func (o *ObservableMap[K, V]) Delete(ctx context.Context, key K) error {
	delete(o.entries, key)
	event := MapChangeEvent[K, V]{
		Source:  o,
		Changed: key,
	}
	return o.OnChange.Emit(ctx, event)
}

func (o *ObservableMap[K, V]) Insert(ctx context.Context, key K, value V) error {
	o.entries[key] = value
	event := MapChangeEvent[K, V]{
		Source:  o,
		Changed: key,
	}
	return o.OnChange.Emit(ctx, event)
}

func (o *ObservableMap[K, V]) Get(ctx context.Context, key K) (V, bool) {
	value, has := o.entries[key]
	return value, has
}

func (o *ObservableMap[K, V]) OnKeyChange(ctx context.Context, key K, onChange emitter.Listener[MapChangeEvent[K, V]]) *emitter.Subscription[MapChangeEvent[K, V]] {
	return o.OnChange.On(func(ctx context.Context, event MapChangeEvent[K, V]) {
		if event.Changed == key {
			onChange(ctx, event)
		}
	})
}
