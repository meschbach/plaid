package mock

import "git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"

// todo: document usage
type engineState struct {
	controller *resources.Client
	procs      *resources.MetaContainer[Proc]
}
