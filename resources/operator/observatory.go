package operator

import (
	"context"
	"github.com/meschbach/plaid/resources"
)

// Observatory watches a set of resources for reconciliation and derives state.
type Observatory[Entity any, Status any] struct {
	//OnChange is called an observed resource has changed.
	OnChange func(ctx context.Context) error
	//Reducer is responsible for taking the Status entity from a resource to reduce into the desired observed status.
	Reducer        func(ctx context.Context, entity Entity) Status
	currentlyWatch map[resources.Meta]resources.WatchToken
}

type ObservedStatus[Status any] struct {
	Ref    resources.Meta
	Exists bool
	Status Status
}

func (o *Observatory[E, T]) Update(ctx context.Context, c *resources.Client, w *resources.ClientWatcher, deps []resources.Meta) ([]ObservedStatus[T], error) {
	oldWatches := o.currentlyWatch
	//todo: test
	o.currentlyWatch = map[resources.Meta]resources.WatchToken{}

	results := make([]ObservedStatus[T], len(deps))
	var signal E
	for i, d := range deps {
		//reuse a watch if we have one for this resource
		if oldWatchToken, has := oldWatches[d]; has {
			delete(oldWatches, d)
			o.currentlyWatch[d] = oldWatchToken
		} else {
			token, err := w.OnResource(ctx, d, func(ctx context.Context, changed resources.ResourceChanged) error {
				if changed.Operation != resources.StatusUpdated {
					return nil
				}
				if o.OnChange != nil {
					return o.OnChange(ctx)
				} else {
					return nil
				}
			})
			if err != nil { //todo: leaking watches -- need to think through
				return results, err
			}
			o.currentlyWatch[d] = token
		}

		exists, err := c.GetStatus(ctx, d, &signal)
		if err != nil { //todo: assuming whole thing is hosed -- and need to be torn down
			return results, err
		}
		if exists {
			results[i] = ObservedStatus[T]{
				Ref:    d,
				Exists: true,
				Status: o.Reducer(ctx, signal),
			}
		} else {
			results[i] = ObservedStatus[T]{
				Ref:    d,
				Exists: false,
			}
		}
	}

	//delete all old references
	for _, token := range oldWatches {
		if err := w.Off(ctx, token); err != nil {
			return results, err
		}
	}
	return results, nil
}

func (o *Observatory[E, T]) Reconcile(ctx context.Context, c *resources.Client) ([]ObservedStatus[T], error) {
	set := make([]ObservedStatus[T], 0, len(o.currentlyWatch))
	var e E
	for d := range o.currentlyWatch {
		exists, err := c.GetStatus(ctx, d, &e)
		if err != nil {
			return set, err
		}
		if !exists {
			set = append(set, ObservedStatus[T]{
				Ref:    d,
				Exists: false,
			})
		} else {
			status := o.Reducer(ctx, e)
			set = append(set, ObservedStatus[T]{
				Ref:    d,
				Exists: true,
				Status: status,
			})
		}
	}
	return set, nil
}
