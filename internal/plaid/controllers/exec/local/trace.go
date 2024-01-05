package local

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("plaid.controllers.exec.local")
