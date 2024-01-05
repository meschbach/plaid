package logging

import (
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/controllers/logdrain"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"git.meschbach.com/mee/platform.git/plaid/ipc/grpc/logger"
	"github.com/thejerf/suture/v4"
)

func NewV1GPRCBridge(supervisor *suture.Supervisor, storageController *resources.Controller, drainConfig *logdrain.ServiceConfig) logger.V1Server {
	wrapper, service := logger.ExportV1LogDrain(storageController, drainConfig)
	supervisor.Add(wrapper)
	return service
}
