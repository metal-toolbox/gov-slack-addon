package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// ErrInvalidTracingConfig is returned when the tracing configuration is invalid.
var ErrInvalidTracingConfig = errors.New("invalid tracing config")

func initTracing() {
	if viper.GetBool("tracing.enabled") {
		initTracer()
	}
}

// initTracer returns an OpenTelemetry TracerProvider.
func initTracer() *tracesdk.TracerProvider {
	exp, err := newExporter()
	if err != nil {
		logger.Fatalw("failed to initialize tracing exporter", "error", err)
	}

	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in an Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("gov-permissions-api-addon"),
			attribute.String("environment", viper.GetString("tracing.environment")),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp
}

func newExporter() (tracesdk.SpanExporter, error) {
	switch viper.GetString("tracing.provider") {
	case "otlpgrpc":
		return newOTLPGRPCExporter()
	case "otlphttp":
		return newOTLPHTTPExporter()
	}

	return nil, fmt.Errorf("%w: tracing exporter must be otlpgrpc or otlphttp", ErrInvalidTracingConfig)
}

func newOTLPGRPCExporter() (tracesdk.SpanExporter, error) {
	_, err := url.Parse(viper.GetString("tracing.endpoint"))
	if err != nil {
		return nil, fmt.Errorf("%w: missing OTLP config options; you must pass a valid endpoint: %w", ErrInvalidTracingConfig, err)
	}

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(viper.GetString("tracing.endpoint")),
		otlptracegrpc.WithTimeout(viper.GetDuration("tracing.timeout")),
	}

	if viper.GetBool("tracing.insecure") {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	return otlptrace.New(context.Background(), otlptracegrpc.NewClient(opts...))
}

func newOTLPHTTPExporter() (tracesdk.SpanExporter, error) {
	_, err := url.Parse(viper.GetString("tracing.endpoint"))
	if err != nil {
		return nil, fmt.Errorf("%w: missing OTLP config options; you must pass a valid endpoint: %w", ErrInvalidTracingConfig, err)
	}

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(viper.GetString("tracing.endpoint")),
		otlptracehttp.WithTimeout(viper.GetDuration("tracing.timeout")),
	}

	if viper.GetBool("tracing.insecure") {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	return otlptrace.New(context.Background(), otlptracehttp.NewClient(opts...))
}
