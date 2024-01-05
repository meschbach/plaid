package daemon

import "git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"

func ExportResources(storage *resources.Client) *ResourceService {
	service := &ResourceService{
		client: storage,
	}
	return service
}
