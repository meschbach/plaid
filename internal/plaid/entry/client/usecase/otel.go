package usecase

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("plaid.client.scenarios")
