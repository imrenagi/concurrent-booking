package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/hibiken/asynq"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/imrenagi/concurrent-booking/booking/handler"
	"github.com/imrenagi/concurrent-booking/booking/services"
	"github.com/imrenagi/concurrent-booking/booking/stores"
	tmetric "github.com/imrenagi/concurrent-booking/internal/telemetry/metric"
	ttrace "github.com/imrenagi/concurrent-booking/internal/telemetry/trace"
	"github.com/imrenagi/concurrent-booking/server/middleware"
)

var name = "booking-service"

type BookingHandler interface {
	Booking() http.HandlerFunc
	BookingV2() http.HandlerFunc
}

// NewServer ...
func NewServer() *Server {

	db := db()

	otelAgentAddr, ok := os.LookupEnv("OTEL_RECEIVER_OTLP_ENDPOINT")
	if !ok {
		log.Fatal().Msg("OTEL_RECEIVER_OTLP_ENDPOINT is not set")
	}

	asynqRedisHost, ok := os.LookupEnv("ASYNQ_REDIS_HOST")
	if !ok {
		log.Fatal().Msg("ASYNC_REDIS_HOST is not set")
	}

	bookingService := services.Booking{
		BookingRepository: stores.NewOrder(db),
		ShowRepository:    stores.NewShow(db),
		Dispatcher:        asynq.NewClient(asynq.RedisClientOpt{Addr: asynqRedisHost}),
	}

	srv := &Server{
		Router:         mux.NewRouter(),
		db:             db,
		bookingHandler: &handler.Handler{Service: bookingService},
	}

	srv.InitGlobalProvider(name, otelAgentAddr)
	srv.routes()

	return srv
}

type Server struct {
	Router *mux.Router
	db     *gorm.DB

	metricProviderCloseFn []tmetric.CloseFunc
	traceProviderCloseFn  []ttrace.CloseFunc

	bookingHandler BookingHandler
}

// Run ...
func (s *Server) Run(ctx context.Context, port int) error {

	httpS := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: s.cors().Handler(middleware.RequestID(s.Router)),
	}

	log.Info().Msgf("server serving on port %d ", port)

	go func() {
		if err := httpS.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msgf("listen:%+s\n", err)
		}
	}()

	<-ctx.Done()

	log.Printf("server stopped")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		cancel()
	}()

	err := httpS.Shutdown(ctxShutDown)
	if err != nil {
		log.Fatal().Msgf("server Shutdown Failed:%+s", err)
	}

	log.Printf("server exited properly")

	if err == http.ErrServerClosed {
		err = nil
	}

	sql, err := s.db.DB()
	if err != nil {
		log.Fatal().Msgf("unable to get db driver")
	}

	if err = sql.Close(); err != nil {
		log.Fatal().Msgf("unable close db connection")
	}

	for _, closeFn := range s.metricProviderCloseFn {
		go func() {
			err = closeFn(ctxShutDown)
			if err != nil {
				log.Error().Err(err).Msgf("Unable to close metric provider")
			}
		}()
	}
	for _, closeFn := range s.traceProviderCloseFn {
		go func() {
			err = closeFn(ctxShutDown)
			if err != nil {
				log.Error().Err(err).Msgf("Unable to close trace provider")
			}
		}()
	}

	return err
}

func (s *Server) cors() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins:     []string{"*"},
		AllowedMethods:     []string{"POST", "GET", "PUT", "DELETE", "HEAD", "OPTIONS"},
		AllowedHeaders:     []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"},
		MaxAge:             60, // 1 minutes
		AllowCredentials:   true,
		OptionsPassthrough: false,
		Debug:              false,
	})
}
