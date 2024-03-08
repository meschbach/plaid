package resources

import "context"

type DirectControllerSystem struct {
	controller *Controller
}

func SystemWithController(c *Controller) *DirectControllerSystem {
	return &DirectControllerSystem{controller: c}
}

func (d *DirectControllerSystem) Storage(ctx context.Context) (Storage, error) {
	return d.controller.Client(), nil
}
