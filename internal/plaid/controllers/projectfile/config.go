package projectfile

import (
	"encoding/json"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/internal/plaid/controllers/probes"
	"github.com/meschbach/plaid/internal/plaid/controllers/project"
	"github.com/meschbach/plaid/internal/plaid/httpProbe"
	"github.com/meschbach/plaid/resources"
)

type Configuration struct {
	//OneShot indicates the service is done once the program is complete.  If the top level project this will result in
	//shutting down the system.
	OneShot  *bool                   `json:"one-shot,omitempty"`
	Name     string                  `json:"name"`
	Run      string                  `json:"run"`
	Build    *BuildConfiguration     `json:"build,omitempty"`
	Requires []string                `json:"requires"`
	Liveness *LivenessConfiguration  `json:"liveness"`
	Ports    *map[string]json.Number `json:"ports"`
}

func (c Configuration) IsOneShot() bool {
	return c.OneShot != nil && *c.OneShot
}

func (c Configuration) toServiceConfig(workingDirectory string) (project.Alpha1DaemonSpec, error) {
	d := project.Alpha1DaemonSpec{
		Name: c.Name,

		Run: exec.TemplateAlpha1Spec{
			Command:    c.Run,
			WorkingDir: workingDirectory,
		},
	}
	if c.Build != nil {
		d.Build = &exec.TemplateAlpha1Spec{
			Command:    c.Build.Exec,
			WorkingDir: workingDirectory,
		}
	}
	for _, dep := range c.Requires {
		d.Requires = append(d.Requires, resources.Meta{Type: project.Alpha1, Name: dep})
	}
	if c.Liveness != nil {
		port, err := (*c.Ports)["http"].Int64()
		if err != nil {
			return d, err
		}
		d.Readiness = &probes.TemplateAlpha1Spec{Http: &httpProbe.TemplateAlpha1{
			Port: uint16(port),
			Path: c.Liveness.Http,
		}}
	}
	return d, nil
}

type LivenessConfiguration struct {
	Http string `json:"http"`
}

type BuildConfiguration struct {
	Exec  string   `json:"exec"`
	Watch []string `json:"watch"`
}
