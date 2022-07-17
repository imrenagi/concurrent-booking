package server

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/unit"
)

func (s *Server) routes() {
	// healthcheck
	s.Router.HandleFunc("/", s.healthcheckHandler)
	s.Router.HandleFunc("/healthz", s.healthcheckHandler)
	s.Router.HandleFunc("/readyz", s.readinessHandler)

	// serve api
	api := s.Router.PathPrefix("/api/v1/").Subrouter()
	api.Use(
		// otelmux this is specific for otlp tracer
		otelmux.Middleware(name),
	)
	api.Handle("/booking", otelhttp.NewHandler(s.bookingHandler.Booking(), "/api/v1/booking"))

	apiV2 := s.Router.PathPrefix("/api/v2/").Subrouter()
	apiV2.Use(
		otelmux.Middleware(name),
	)
	apiV2.Handle("/booking", otelhttp.NewHandler(s.bookingHandler.BookingV2(), "/api/v2/booking"))
}

var meter = global.Meter("ex.com/basic")
var lemonsKey = attribute.Key("ex.com/lemons")
var commonAttrs = []attribute.KeyValue{lemonsKey.Int(10), attribute.String("A", "1"), attribute.String("B", "2"), attribute.String("C", "3")}

func (s *Server) healthcheckHandler(w http.ResponseWriter, r *http.Request) {

	counter, _ := meter.SyncFloat64().Counter("server.healthcheck",
		instrument.WithDescription("testing healthcheck count"),
		instrument.WithUnit(unit.Dimensionless),
	)
	counter.Add(r.Context(), 1.0, commonAttrs...)

	w.Write([]byte("im alive"))
}

func (s *Server) readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("im ready to face the world"))
}