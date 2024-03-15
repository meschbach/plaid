package alpha2

import (
	"github.com/meschbach/plaid/internal/plaid/controllers/dependencies"
	"github.com/meschbach/plaid/resources"
	"time"
)

type Status struct {
	//LatestToken seen by the controller and working towards propagating
	LatestToken string `json:"latest-token"`
	//Debounce    time.Duration `json:"debounce"`
	//Stable has been built and is ready for service
	Stable *TokenStatus `json:"stable,omitempty"`
	//Next is an instance of the system which is building, start, or readying.  Once ready it will dispatch
	Next *TokenStatus `json:"next,omitempty"`
	//Old token which are to be retired
	Old []TokenStatus `json:"old,omitempty"`
	//Ready indicates the current stable service matches LatestToken and is stable
	Ready bool `json:"ready"`
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
	//Ready indicates the service is running and considered ready
	Ready bool `json:"ready"`
	//Deps encapsulates the state of the dependencies
	Deps dependencies.Alpha1Status `json:"dependencies,omitempty"`
	//DepsFuse represents if all dependencies have been seen as available
	DepsFuse bool `json:"dependencies-fuse"`
}
