package exporter

import (
	"context"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
)

func NewOTLP(endpoint string) *otlpmetric.Exporter {
	ctx := context.Background()
	metricClient := otlpmetricgrpc.NewClient(
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(endpoint))

	metricExp, err := otlpmetric.New(ctx, metricClient)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to create the collector metric exporter")
	}

	return metricExp
}
