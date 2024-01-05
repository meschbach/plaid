package logging

import (
	"github.com/meschbach/plaid/internal/plaid/controllers/logdrain"
	"github.com/meschbach/plaid/ipc/grpc/logger"
	"github.com/meschbach/plaid/resources"
	"github.com/thejerf/suture/v4"
)

func NewV1GPRCBridge(supervisor *suture.Supervisor, storageController *resources.Controller, drainConfig *logdrain.ServiceConfig) logger.V1Server {
	wrapper, service := logger.ExportV1LogDrain(storageController, drainConfig)
	supervisor.Add(wrapper)
	return service
}
