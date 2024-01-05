package project

import (
	"github.com/meschbach/plaid/internal/plaid/controllers/dependencies"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/internal/plaid/controllers/probes"
	"github.com/meschbach/plaid/internal/plaid/resources"
)

const Kind = "plaid.meschbach.com/project"

var Alpha1 = resources.Type{
	Kind:    Kind,
	Version: "alpha1",
}

type Alpha1Spec struct {
	BaseDirectory string              `json:"base-directory"`
	OneShots      []Alpha1OneShotSpec `json:"one-shots,omitempty"`
	Daemons       []Alpha1DaemonSpec  `json:"daemons,omitempty"`
}

type Alpha1OneShotSpec struct {
	Name     string                  `json:"name"`
	Build    exec.TemplateAlpha1Spec `json:"build"`
	Run      exec.TemplateAlpha1Spec `json:"run"`
	Requires dependencies.Alpha1Spec `json:"requires,omitempty"`
}

type Alpha1DaemonSpec struct {
	Name      string                     `json:"name"`
	Build     *exec.TemplateAlpha1Spec   `json:"build,omitempty"`
	Run       exec.TemplateAlpha1Spec    `json:"run"`
	Requires  []resources.Meta           `json:"requires,omitempty"`
	Readiness *probes.TemplateAlpha1Spec `json:"readiness,omitempty"`
}

type Alpha1Status struct {
	OneShots []Alpha1OneShotStatus `json:"one-shots,omitempty"`
	Daemons  []*Alpha1DaemonStatus `json:"daemons,omitempty"`
	//Done flags when all one-shot type builds have completed
	Done   bool   `json:"done"`
	Result string `json:"result"`
	Ready  bool   `json:"ready"`
}

const Alpha1StateCreating = "creating"
const Alpha1StateProgressing = "progressing"
const Alpha1StateSuccess = "success"
const Alpha1StateFailed = "failed"

type Alpha1OneShotStatus struct {
	Name  string
	Done  bool            `json:"done"`
	State string          `json:"state"`
	Ref   *resources.Meta `json:"ref,omitempty"`
}

type Alpha1DaemonStatus struct {
	Name    string          `json:"name,omitempty"`
	Current *resources.Meta `json:"current,omitempty"`
	Ready   bool            `json:"ready"`
}
