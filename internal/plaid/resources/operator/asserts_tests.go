package operator

import (
	"context"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"testing"
)

func AssertHealthy(t *testing.T, ctx context.Context, res *resources.Client, ref resources.Meta) {
	resources.AssertStatus(t, ctx, res, ref, func(status HealthySignal) bool {
		return status.Healthy
	})
}
