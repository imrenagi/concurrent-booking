package metric

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/metric/export"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
)

type CloseFunc func(ctx context.Context) error

func NewMeterProviderBuilder() *meterProviderBuilder {
	return &meterProviderBuilder{}
}

type meterProviderBuilder struct {
	exporter            export.Exporter
	histogramBoundaries []float64
}

func (b *meterProviderBuilder) SetExporter(exp export.Exporter) *meterProviderBuilder {
	b.exporter = exp
	return b
}

func (b *meterProviderBuilder) SetHistogramBoundaries(explicitBoundaries []float64) *meterProviderBuilder {
	b.histogramBoundaries = explicitBoundaries
	return b
}

func (b meterProviderBuilder) Build() (metric.MeterProvider, CloseFunc, error) {
	var histogramOptions []histogram.Option
	if len(b.histogramBoundaries) > 0 {
		histogramOptions = append(histogramOptions, histogram.WithExplicitBoundaries(b.histogramBoundaries))
	}

	if b.exporter == nil {
		return nil, nil, fmt.Errorf("exporter is not set")
	}

	pusher := controller.New(
		processor.NewFactory(
			simple.NewWithHistogramDistribution(histogramOptions...),
			b.exporter,
		),
		controller.WithExporter(b.exporter),
		controller.WithCollectPeriod(5*time.Second),
	)

	if err := pusher.Start(context.Background()); err != nil {
		return nil, nil, err
	}

	return pusher, func(ctx context.Context) error {
		cxt, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := pusher.Stop(cxt); err != nil {
			return err
		}
		return nil
	}, nil
}
