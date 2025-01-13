package otel

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

const (
	SpanKingCPU     = "cpu"
	SpanKindDB      = "db"
	SpanKindRPC     = "rpc"
	SpanKindHTTP    = "http"
	SpanKindRequest = "request"
)

func SetError(span trace.Span, err error) {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}
}
func Start(ctx context.Context, name string, spanKind string) (context.Context, trace.Span) {
	if spanKind == "" {
		spanKind = SpanKingCPU
	}
	ctx, spn := tracer.Start(ctx, name)
	spn.SetAttributes(attribute.String("span.kind", spanKind))
	return ctx, spn
}
