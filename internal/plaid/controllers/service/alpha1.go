package service

import (
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/internal/plaid/controllers/probes"
	"github.com/meschbach/plaid/resources"
)

var Alpha1 = resources.Type{
	Kind:    Kind,
	Version: "alpha1",
}

type Alpha1Spec struct {
	// Dependencies are checked for the status with a `ready` field.  If all dependencies are ready then the a build and
	// run is issued
	Dependencies []resources.Meta         `json:"dependencies,omitempty"`
	Build        *exec.TemplateAlpha1Spec `json:"build,omitempty"`
	Run          exec.TemplateAlpha1Spec  `json:"run"`
	// Readiness defines a probe to determine if the system is ready
	Readiness *probes.TemplateAlpha1Spec `json:"readiness,omitempty"`
	// When RestartToken changes a new build will occur (if one exists) and if successful will stop the existing process
	// and start the new one.
	RestartToken string `json:"restart-token"`
}

type Alpha1Status struct {
	Dependencies []Alpha1StatusDependency `json:"dependencies,omitempty"`
	Build        Alpha1BuildStatus        `json:"build,omitempty"`
	Run          Alpha1RunStatus          `json:"run"`
	Ready        bool                     `json:"ready"`
	RunningToken string                   `json:"running-token"`
}

type Alpha1StatusDependency struct {
	Dependency resources.Meta `json:"ref"`
	Ready      bool           `json:"ready"`
}

type Alpha1BuildStatus struct {
	State string          `json:"state"`
	Ref   *resources.Meta `json:"ref,omitempty"`
}

const StateNotReady = "not-ready"
const Running = "running"

type Alpha1RunStatus struct {
	State string          `json:"state"`
	Ref   *resources.Meta `json:"ref,omitempty"`
}
