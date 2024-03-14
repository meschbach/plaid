package project

import (
	"context"
	"github.com/go-faker/faker/v4"
	"github.com/meschbach/plaid/controllers/service/alpha2"
	"github.com/meschbach/plaid/controllers/tooling"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/optest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func envFrom(plaid *optest.System, subject resources.Meta) tooling.Env {
	return tooling.Env{
		Subject: subject,
		Storage: plaid.Storage,
		Watcher: plaid.Observer,
		Reconcile: func(ctx context.Context) error {
			return nil
		},
	}
}

func TestDaemonState(t *testing.T) {
	t.Run("Given a new project space", func(t *testing.T) {
		_, plaid := optest.New(t)

		exampleProject := resources.Meta{Type: Alpha1, Name: faker.Name()}

		daemon := &daemonState{}

		exampleDaemonSpec := Alpha1DaemonSpec{}
		exampleSpec := Alpha1Spec{}

		plaid.Run("When asked for next steps", func(t *testing.T, plaid *optest.System, ctx context.Context) {
			step, err := daemon.decideNextStep(ctx, envFrom(plaid, exampleProject))
			require.NoError(t, err)
			assert.Equal(t, daemonCreate, step, "Then we will create our resources")
		})

		plaid.Run("When Created", func(t *testing.T, plaid *optest.System, ctx context.Context) {
			err := daemon.create(ctx, envFrom(plaid, exampleProject), exampleSpec, exampleDaemonSpec)
			require.NoError(t, err)
			step, err := daemon.decideNextStep(ctx, envFrom(plaid, exampleProject))
			require.NoError(t, err)

			assert.Equal(t, daemonWait, step, "Then the next step is wait, got %s", step)
			status := &Alpha1DaemonStatus{}
			daemon.toStatus(exampleDaemonSpec, status)
			assert.Equal(t, daemon.service.Ref, *status.Current, "Then reports correct service Ref")
			assert.False(t, status.Ready, "Then it is not ready")
		})

		plaid.Run("When daemon has become ready", func(t *testing.T, plaid *optest.System, ctx context.Context) {
			plaid.MustUpdateStatus(ctx, daemon.service.Ref, alpha2.Status{
				LatestToken: "",
				Ready:       true,
				Stable: &alpha2.TokenStatus{
					Token: "",
					Ready: true,
				},
			})

			step, err := daemon.decideNextStep(ctx, envFrom(plaid, exampleProject))
			require.NoError(t, err)

			assert.Equal(t, daemonWait, step, "Then the next step is to wait")
			status := &Alpha1DaemonStatus{}
			daemon.toStatus(exampleDaemonSpec, status)
			assert.Equal(t, daemon.service.Ref, *status.Current, "Then reports correct service Ref")
			assert.True(t, status.Ready, "Then it is ready")
		})
	})
}
