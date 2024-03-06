package service

import "github.com/meschbach/plaid/resources"

func New(export *resources.Client) *ResourceService {
	return &ResourceService{
		client: export,
	}
}
