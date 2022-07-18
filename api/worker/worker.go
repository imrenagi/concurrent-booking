package worker

import (
	"context"
	"os"
	"time"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"

	tmetric "github.com/imrenagi/concurrent-booking/api/internal/telemetry/metric"
	ttrace "github.com/imrenagi/concurrent-booking/api/internal/telemetry/trace"
	"github.com/imrenagi/concurrent-booking/api/worker/middleware"
)

var name = "reservation-worker"
var trc = otel.Tracer("github.com/imrenagi/concurrent-booking/api/cmd/worker")

func NewWorker() *Worker {

	asynqRedisHost, ok := os.LookupEnv("ASYNQ_REDIS_HOST")
	if !ok {
		log.Fatal().Msg("ASYNC_REDIS_HOST is not set")
	}

	otelAgentAddr, ok := os.LookupEnv("OTEL_RECEIVER_OTLP_ENDPOINT")
	if !ok {
		log.Fatal().Msg("OTEL_RECEIVER_OTLP_ENDPOINT is not set")
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: asynqRedisHost},
		asynq.Config{
			Concurrency: 100,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	w := &Worker{
		// Router: asynq.NewServeMux(),
		server:  srv,
		mux:     asynq.NewServeMux(),
		handler: newHandler(),
	}

	w.routes()
	w.InitGlobalProvider(name, otelAgentAddr)

	return w
}

type Worker struct {
	mux     *asynq.ServeMux
	server  *asynq.Server
	handler *handler

	metricProviderCloseFn []tmetric.CloseFunc
	traceProviderCloseFn  []ttrace.CloseFunc
}

func (w *Worker) routes() {
	w.mux.Use(middleware.SpanPropagator)
	w.mux.HandleFunc("booking:reserve", w.handler.HandleReservation)

}

func (w *Worker) Run(ctx context.Context) error {

	go func() {
		if err := w.server.Run(w.mux); err != nil {
			log.Fatal().Msgf("could not run server: %v", err)
		}
	}()

	<-ctx.Done()

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		cancel()
	}()

	w.server.Stop()
	w.server.Shutdown()

	for _, closeFn := range w.metricProviderCloseFn {
		go func() {
			err := closeFn(ctxShutDown)
			if err != nil {
				log.Error().Err(err).Msgf("Unable to close metric provider")
			}
		}()
	}
	for _, closeFn := range w.traceProviderCloseFn {
		go func() {
			err := closeFn(ctxShutDown)
			if err != nil {
				log.Error().Err(err).Msgf("Unable to close trace provider")
			}
		}()
	}

	return nil
}
