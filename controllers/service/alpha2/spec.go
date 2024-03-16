package alpha2

import (
	"github.com/meschbach/plaid/internal/plaid/controllers/dependencies"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/internal/plaid/controllers/probes"
)

type Spec struct {
	// Dependencies are checked for the status with a `ready` field.  If all dependencies are ready then the a build and
	// run is issued
	Dependencies dependencies.Alpha1Spec  `json:"dependencies,omitempty"`
	Build        *exec.TemplateAlpha1Spec `json:"build,omitempty"`
	Run          exec.TemplateAlpha1Spec  `json:"run"`
	// Readiness defines a probe to determine if the system is ready
	Readiness *probes.TemplateAlpha1Spec `json:"readiness,omitempty"`
	// When RestartToken changes a new build will occur (if one exists) and if successful will stop the existing process
	// and start the new one.
	RestartToken string `json:"restart-token"`
}
