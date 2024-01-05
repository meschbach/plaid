package projectfile

import (
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDaemonConfigSetup(t *testing.T) {
	t.Run("Given a service configuration without a build", func(t *testing.T) {
		cfg := Configuration{
			Name: faker.Name(),
			Run:  "./dev.sh",
		}
		wd := "/wc/some-work"

		t.Run("When building the service configuration", func(t *testing.T) {
			res, err := cfg.toServiceConfig(wd)
			require.NoError(t, err)
			assert.Nil(t, res.Build, "Then no build configuration is setup")
		})
	})
}
