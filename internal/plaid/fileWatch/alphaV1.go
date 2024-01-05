package fileWatch

import (
	"github.com/meschbach/plaid/internal/plaid/resources"
	"time"
)

const Kind = "file-watch.plaid.meschbach.com"

var AlphaV1Type = resources.Type{
	Kind:    Kind,
	Version: "alphaV1",
}

type AlphaV1Spec struct {
	//AbsolutePath is the root file system location to observe
	AbsolutePath string `json:"absolute-path"`
	//Recursive indicates the watch would like to monitor the target directory and all subdirectories.
	Recursive bool
}

type AlphaV1Status struct {
	Watching   bool       `json:"watching,omitempty"`
	LastChange *time.Time `json:"last-change,omitempty"`
}
