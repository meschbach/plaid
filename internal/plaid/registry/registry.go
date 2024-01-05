package registry

import (
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"time"
)

// registry represents the state of a single registry set
type registry struct {
	loaded   *time.Time
	config   Config
	projects map[string]resources.Meta
}
