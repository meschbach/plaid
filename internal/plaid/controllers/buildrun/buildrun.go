// Package buildrun provides a resource and associated controllers to (re)build and (re)start a particular program for
// execution.  Restarts are signaled through the change of a restart token which by default is an empty string.
package buildrun

import (
	"github.com/meschbach/plaid/internal/plaid/controllers/dependencies"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/resources"
)

const Kind = "plaid.meschbach.com/buildrun"

var Alpha1 = resources.Type{
	Kind:    Kind,
	Version: "alpha1",
}

type AlphaSpec1 struct {
	RestartToken string                  `json:"restart-token"`
	Build        exec.TemplateAlpha1Spec `json:"build,omitempty"`
	Run          exec.TemplateAlpha1Spec `json:"run"`
	Requires     dependencies.Alpha1Spec `json:"requires,omitempty"`
}

type AlphaStatus1 struct {
	Build        Alpha1StatusBuild         `json:"build"`
	Run          Alpha1StatusRun           `json:"run"`
	Dependencies dependencies.Alpha1Status `json:"dependencies"`
	Ready        bool                      `json:"ready"`
}

type Alpha1StatusBuild struct {
	//Token is the restart-token last used when attempting to run the build
	Token  string                        `json:"token"`
	Result *exec.InvocationAlphaV1Status `json:"result"`
	State  string                        `json:"state"`
	Ref    *resources.Meta               `json:"ref"`
}

type Alpha1StatusRun struct {
	//Token is the restart-token last used when attempting to run the build
	Token  string                        `json:"token"`
	State  string                        `json:"state"`
	Result *exec.InvocationAlphaV1Status `json:"result"`
	Ref    *resources.Meta               `json:"ref"`
}
