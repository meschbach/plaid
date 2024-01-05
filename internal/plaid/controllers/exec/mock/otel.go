package mock

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("plaid.controllers.exec.mock")
