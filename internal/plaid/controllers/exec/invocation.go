package exec

import (
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"time"
)

// InvocationKind represents a single invocation of a local process to the running controller.
const InvocationKind = "shell-invocation.plaid.meschbach.com"

var InvocationAlphaV1Type = resources.Type{
	Kind:    InvocationKind,
	Version: "alphaV1",
}

type InvocationAlphaV1Spec struct {
	//Exec is the string to be parsed and executed.  Arguments are space separated without regards for quotes.
	Exec string `json:"exec"`
	//WorkingDir is the working directory to launch the process in.
	WorkingDir string `json:"working-dir"`
}

type InvocationAlphaV1Status struct {
	//TODO
	//PID        *string    `json:"pid,omitempty"`
	Started  *time.Time `json:"started,omitempty"`
	Finished *time.Time `json:"finished,omitempty"`
	//TODO
	ExitStatus *int `json:"exit-status,omitempty"`
	Healthy    bool `json:"healthy"`
}
