package daemon

import "github.com/meschbach/plaid/resources"

func ExportResources(storage *resources.Client) *ResourceService {
	service := &ResourceService{
		client: storage,
	}
	return service
}
