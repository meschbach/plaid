package mock

import "github.com/meschbach/plaid/internal/plaid/resources"

// todo: document usage
type engineState struct {
	controller *resources.Client
	procs      *resources.MetaContainer[Proc]
}
