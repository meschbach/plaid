package logger

import (
	"go.opentelemetry.io/otel"
)

var tracing = otel.Tracer("git.meschbach.com/mee/platform.git/plaid/ipc/grpc/logger")
