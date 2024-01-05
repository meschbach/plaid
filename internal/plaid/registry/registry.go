package registry

import (
	"github.com/meschbach/plaid/internal/plaid/resources"
	"time"
)

// registry represents the state of a single registry set
type registry struct {
	loaded   *time.Time
	config   Config
	projects map[string]resources.Meta
}
