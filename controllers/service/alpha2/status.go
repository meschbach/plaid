package alpha2

import (
	"github.com/meschbach/plaid/resources"
	"time"
)

type Status struct {
	LatestToken string `json:"latest-token"`
	//Debounce    time.Duration `json:"debounce"`
	//Stable has been built and is ready for service
	Stable *TokenStatus `json:"stable,omitempty"`
	//Next is an instance of the system which is building, start, or readying.  Once ready it will dispatch
	Next *TokenStatus `json:"next,omitempty"`
	//Old token which are to be retired
	Old []TokenStatus `json:"old,omitempty"`
}

const TokenStageInit = "init"
const TokenStageDependencyWait = "dep-wait"
const TokenStageDependenciesReady = "dep-ready"
const TokenStageBuilding = "building"
const TokenStageStarting = "starting"
const TokenStageReady = "ready"
const TokenStageStopping = "stopping"
const TokenStageStopped = "stopped"

type TokenStatus struct {
	Token string `json:"token"`
	Stage string `json:"stage"`
	//Last time when a change has occurred to this resource
	Last time.Time `json:"last"`
	//Service refers to the process which is the service.  Will be nil until the service is created.
	Service *resources.Meta `json:"service,omitempty"`
}
