package service

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("plaid.daemon.grpc-v1/service")
