package registry

import "github.com/meschbach/plaid/resources"

type AlphaV1Spec struct {
	AbsoluteFilePath string `json:"absolute-file-path"`
}

type AlphaV1Status struct {
	Problem string `json:"problem,omitempty"`
}

const Kind = "registry.plaid.meschbach.com"

var AlphaV1 = resources.Type{
	Kind:    Kind,
	Version: "alpha1",
}
