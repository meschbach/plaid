package project

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("plaid.controllers.project")
