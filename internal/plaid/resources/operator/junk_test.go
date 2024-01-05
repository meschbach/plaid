package operator

import (
	"context"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"testing"
)

type plaidHarness struct {
	core            *resources.TestSubsystem
	storeController *resources.Controller
	store           *resources.Client
	systemDone      func()
}

func withPlaidTest(t *testing.T, ctx context.Context) *plaidHarness {
	core := resources.WithTestSubsystem(t, ctx)

	return &plaidHarness{
		storeController: core.Controller,
		store:           core.Store,
		systemDone: func() {
			core.SystemDone()
		},
	}
}
