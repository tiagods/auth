package otel

import (
	"context"
	"errors"
	"github.com/tiagods/auth/internal/infra/env"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"time"
)

// SetupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error
	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}
	// Set up trace provider.
	tracerProvider, err := newTraceProvider(ctx)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	tracer = tracerProvider.Tracer(env.GetEnvAsString(env.SERVICE_NAME, env.DEFAULT_SERVICE_NAME))
	return
}

func newTraceProvider(ctx context.Context) (*sdktrace.TracerProvider, error) {
	// Ensure default SDK resources and the required service name are set.
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(env.GetEnvAsString(env.SERVICE_NAME, env.DEFAULT_SERVICE_NAME)),
			semconv.ServiceVersion(env.GetEnvAsString(env.VERSION, env.DEFAULT_VERSION)),
		),
	)

	tracerExp, err := otlptracehttp.New(ctx, otlptracehttp.WithInsecure(), otlptracehttp.WithEndpoint(env.GetEnvAsString(env.OTEL_EXPORTER_OTLP_ENDPOINT, env.DEFAULT_OTEL_EXPORTER_OTLP_ENDPOINT)))
	if err != nil {
		return nil, err
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(tracerExp, sdktrace.WithBatchTimeout(time.Second)),
		sdktrace.WithResource(r),
	), nil
}
