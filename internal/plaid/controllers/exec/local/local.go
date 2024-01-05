package local

import (
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/controllers/logdrain"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"github.com/thejerf/suture/v4"
)

func NewSystem(resources *resources.Controller, logging *logdrain.ServiceConfig, procSupervisors *suture.Supervisor) *Controller {
	return &Controller{
		storage:         resources,
		procSupervisors: procSupervisors,
		logging:         logging,
	}
}
