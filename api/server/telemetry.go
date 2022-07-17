package server

import (
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"

	"github.com/imrenagi/concurrent-booking/api/internal/telemetry/metric"
	metricExporter "github.com/imrenagi/concurrent-booking/api/internal/telemetry/metric/exporter"
	ttrace "github.com/imrenagi/concurrent-booking/api/internal/telemetry/trace"
	traceExporter "github.com/imrenagi/concurrent-booking/api/internal/telemetry/trace/exporter"
)

func (s *Server) InitGlobalProvider(name, endpoint string) {
	metricExp := metricExporter.NewOTLP(endpoint)
	pusher, pusherCloseFn, err := metric.NewMeterProviderBuilder().
		SetExporter(metricExp).
		SetHistogramBoundaries([]float64{5, 10, 25, 50, 100, 200, 400, 800, 1000}).
		Build()
	if err != nil {
		log.Fatal().Err(err).Msgf("failed initializing the meter provider")
	}
	s.metricProviderCloseFn = append(s.metricProviderCloseFn, pusherCloseFn)
	global.SetMeterProvider(pusher)

	spanExporter := traceExporter.NewOTLP(endpoint)
	tracerProvider, tracerProviderCloseFn, err := ttrace.NewTraceProviderBuilder(name).
		SetExporter(spanExporter).
		Build()
	if err != nil {
		log.Fatal().Err(err).Msgf("failed initializing the tracer provider")
	}
	s.traceProviderCloseFn = append(s.traceProviderCloseFn, tracerProviderCloseFn)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otel.SetTracerProvider(tracerProvider)
}
