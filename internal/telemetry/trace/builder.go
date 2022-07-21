package trace

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type CloseFunc func(ctx context.Context) error

func NewTraceProviderBuilder(name string) *traceProviderBuilder {
	return &traceProviderBuilder{
		name: name,
	}
}

type traceProviderBuilder struct {
	name     string
	exporter trace.SpanExporter
}

func (b *traceProviderBuilder) SetExporter(exp trace.SpanExporter) *traceProviderBuilder {
	b.exporter = exp
	return b
}

func (b *traceProviderBuilder) Build() (*trace.TracerProvider, CloseFunc, error) {
	ctx := context.Background()
	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		// resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			// the service name used to display traces in backend
			semconv.ServiceNameKey.String(b.name),
		),
	)
	if err != nil {
		return nil, nil, err
	}

	bsp := trace.NewBatchSpanProcessor(b.exporter)
	tracerProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
		trace.WithSpanProcessor(bsp),
	)

	return tracerProvider, func(ctx context.Context) error {
		cxt, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := b.exporter.Shutdown(cxt); err != nil {
			return err
		}
		return err
	}, nil
}
