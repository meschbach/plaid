package httpProbe

import "github.com/meschbach/plaid/resources"

type AlphaV1Spec struct {
	Enabled  bool   `json:"enabled"`
	Host     string `json:"host"`
	Port     uint16 `json:"port"`
	Resource string `json:"uri"`
}

type AlphaV1Status struct {
	Ready bool `json:"ready"`
}

const Kind = "http-probe.plaid.meschbach.com"

var AlphaV1Type = resources.Type{
	Kind:    Kind,
	Version: "alpha1",
}
