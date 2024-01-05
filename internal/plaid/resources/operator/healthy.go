package operator

import (
	"context"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
)

type HealthyDependenciesMemoization struct {
	OnChange    func(ctx context.Context) error
	observatory Observatory[HealthySignal, HealthySet]
}

func NewHealthDependencies() *HealthyDependenciesMemoization {
	h := &HealthyDependenciesMemoization{
		OnChange:    nil,
		observatory: Observatory[HealthySignal, HealthySet]{},
	}
	h.observatory.OnChange = func(ctx context.Context) error {
		if h.OnChange != nil {
			return h.OnChange(ctx)
		}
		return nil
	}
	h.observatory.Reducer = reduceHealthSignal
	return h
}

func (h *HealthyDependenciesMemoization) Update(ctx context.Context, c *resources.Client, w *resources.ClientWatcher, deps []resources.Meta) (bool, []ObservedStatus[HealthySet], error) {
	status, err := h.observatory.Update(ctx, c, w, deps)
	if err != nil {
		return false, status, err
	}
	allHealthy := true
	for _, h := range status {
		allHealthy = allHealthy && h.Exists && h.Status.Healthy
	}
	return allHealthy, status, nil
}

func (h *HealthyDependenciesMemoization) Reconcile(ctx context.Context, c *resources.Client) (bool, []ObservedStatus[HealthySet], error) {
	status, err := h.observatory.Reconcile(ctx, c)
	if err != nil {
		return false, status, err
	}
	allHealthy := true
	for _, h := range status {
		allHealthy = allHealthy && h.Exists && h.Status.Healthy
	}
	return allHealthy, status, nil
}

type HealthySignal struct {
	Healthy bool `json:"healthy"`
}

type HealthySet struct {
	Healthy bool `json:"healthy"`
}

func reduceHealthSignal(ctx context.Context, in HealthySignal) HealthySet {
	return HealthySet{Healthy: in.Healthy}
}
