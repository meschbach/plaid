package mock

import "github.com/meschbach/plaid/resources"

// todo: document usage
type engineState struct {
	controller *resources.Client
	procs      *resources.MetaContainer[Proc]
}
