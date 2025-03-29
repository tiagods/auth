package tracer

import (
	"context"

	"github.com/tiagods/auth/internal/infra/env"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var serviceName = env.GetEnvAsString(env.SERVICE_NAME, env.DEFAULT_SERVICE_NAME)

const (
	SpanKindCPU       = "cpu"
	SpanKindDB        = "db"
	SpanKindRPC       = "rpc"
	SpanKindHTTP      = "http"
	SpanKindRequest   = "request"
	SpanKindClient    = "client"
	SpanKindServer    = "server"
	SpanKindConsumer  = "consumer"
	SpanKindProducer  = "producer"
	SpanKindInternal  = "internal"
	SpanKindCache     = "cache"
	SpanKindMessaging = "messaging"
)

type SpanKind struct {
	SpanKindCPU string
	Value       trace.SpanKind
}

func SetError(span trace.Span, err error) {
	if err != nil && span != nil {
		span.SetAttributes(
			attribute.String("error.message", err.Error()),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}
func SetWarning(span trace.Span, err error) {
	if err != nil || span != nil {
		span.SetAttributes(attribute.String("warn", err.Error()))
	}
}

func SetAttributes(span trace.Span, attributes ...attribute.KeyValue) {
	span.SetAttributes(attributes...)
}

func AddEvent(span trace.Span, name string, attributes ...attribute.KeyValue) {
	span.AddEvent(name, trace.WithAttributes(attributes...))
}

func Start(ctx context.Context, name string, spanKind string) (context.Context, trace.Span) {
	ctx, spn := otel.Tracer(serviceName).Start(ctx, name, trace.WithSpanKind(getSpanKind(spanKind)),
		trace.WithAttributes(attribute.String("span.kind", spanKind)))
	return ctx, spn
}

func getSpanKind(spanKind string) trace.SpanKind {
	switch spanKind {
	case SpanKindCPU:
		return trace.SpanKindInternal
	case SpanKindDB:
		return trace.SpanKindInternal
	case SpanKindRPC:
		return trace.SpanKindClient
	case SpanKindHTTP:
		return trace.SpanKindClient
	case SpanKindRequest:
		return trace.SpanKindClient
	case SpanKindClient:
		return trace.SpanKindClient
	case SpanKindServer:
		return trace.SpanKindServer
	case SpanKindConsumer:
		return trace.SpanKindConsumer
	case SpanKindProducer:
		return trace.SpanKindProducer
	case SpanKindInternal:
		return trace.SpanKindInternal
	case SpanKindCache:
		return trace.SpanKindClient
	case SpanKindMessaging:
		return trace.SpanKindClient
	default:
		return trace.SpanKindInternal
	}
}
