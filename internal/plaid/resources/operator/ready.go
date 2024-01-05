package operator

import (
	"context"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
)

type ReadySignal struct {
	Ready bool `json:"ready"`
}

type ReadinessObserver struct {
	OnChange    func(ctx context.Context) error
	observatory Observatory[ReadySignal, ReadyStatus]
}

type ReadyStatus struct {
	Ready bool `json:"ready"`
}

func reduceReadySignalToReadyStatus(ctx context.Context, entity ReadySignal) ReadyStatus {
	return ReadyStatus{
		Ready: entity.Ready,
	}
}

func NewReadinessObserver() *ReadinessObserver {
	r := &ReadinessObserver{
		OnChange: nil,
	}
	r.observatory.OnChange = func(ctx context.Context) error {
		if r.OnChange != nil {
			return r.OnChange(ctx)
		}
		return nil
	}
	//todo: pull out
	r.observatory.Reducer = reduceReadySignalToReadyStatus
	return r
}

func (r *ReadinessObserver) Update(ctx context.Context, c *resources.Client, w *resources.ClientWatcher, deps []resources.Meta) (bool, error) {
	observed, err := r.observatory.Update(ctx, c, w, deps)
	if err != nil {
		return false, err
	}
	allReady := true
	for _, o := range observed {
		allReady = allReady && o.Exists && o.Status.Ready
	}
	return allReady, nil
}

func (r *ReadinessObserver) Reconcile(ctx context.Context, c *resources.Client) (bool, []ObservedStatus[ReadyStatus], error) {
	observed, err := r.observatory.Reconcile(ctx, c)
	if err != nil {
		return false, nil, err
	}
	allReady := true
	for _, o := range observed {
		allReady = allReady && o.Exists && o.Status.Ready
	}
	return allReady, observed, nil
}
