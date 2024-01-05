package local

import (
	"github.com/meschbach/plaid/internal/plaid/controllers/logdrain"
	"github.com/meschbach/plaid/resources"
	"github.com/thejerf/suture/v4"
)

func NewSystem(resources *resources.Controller, logging *logdrain.ServiceConfig, procSupervisors *suture.Supervisor) *Controller {
	return &Controller{
		storage:         resources,
		procSupervisors: procSupervisors,
		logging:         logging,
	}
}
