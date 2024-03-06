package kit

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("plaid.controllers.tooling.kit")
