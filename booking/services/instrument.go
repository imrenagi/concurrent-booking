package services

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/unit"
)

var tracer = otel.Tracer("github.com/imrenagi/concurrent-booking/booking/services")

var meter = global.Meter("github.com/imrenagi/concurrent-booking/booking/services")

var orderCounter, _ = meter.SyncInt64().Counter("order",
	instrument.WithDescription("number of order with its status"),
	instrument.WithUnit(unit.Dimensionless))

var orderStatusKey = attribute.Key("status")
