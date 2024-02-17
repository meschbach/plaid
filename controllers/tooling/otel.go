package tooling

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("plaid.controllers.tooling")
