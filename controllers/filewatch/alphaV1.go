package filewatch

import (
	"github.com/meschbach/plaid/resources"
	"time"
)

const Kind = "file-watch.plaid.meschbach.com"

var Alpha1 = resources.Type{
	Kind:    Kind,
	Version: "alpha1",
}

type Alpha1Spec struct {
	//AbsolutePath is the root file system location to observe
	AbsolutePath string `json:"absolute-Path"`
}

type Alpha1Status struct {
	Watching   bool       `json:"watching,omitempty"`
	LastChange *time.Time `json:"last-change,omitempty"`
}
