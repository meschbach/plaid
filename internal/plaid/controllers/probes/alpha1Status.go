package probes

import "github.com/meschbach/plaid/resources"

type TemplateAlpha1Status struct {
	Ref   *resources.Meta `json:"ref"`
	Ready bool            `json:"ready"`
}
