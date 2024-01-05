package resources

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var tracing = otel.Tracer("git.meschbach.com/mee/platform.git/plaid/internal/resources")

// todo: move to junk o11y
func TracedMessageContext(parent context.Context, parentSpan trace.SpanContext, opName string, options ...trace.SpanStartOption) (context.Context, trace.Span) {
	parentTraceContext := trace.ContextWithRemoteSpanContext(parent, parentSpan)
	return tracing.Start(parentTraceContext, opName, options...)
}

func TracedMessageConsumer(parent context.Context, producerContext trace.Link, opName string, options ...trace.SpanStartOption) (context.Context, trace.Span) {
	options = append(options, trace.WithSpanKind(trace.SpanKindConsumer), trace.WithLinks(producerContext))
	return tracing.Start(parent, opName, options...)
}
