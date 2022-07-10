package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"

	// "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/imrenagi/concurrent-booking/api/booking"
	"github.com/imrenagi/concurrent-booking/api/pkg/tracer"
)

// NewServer ...
func NewServer() *Server {

	ctx := context.Background()
	db, err := gormDB(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to start db")
	}

	otelAgentAddr, ok := os.LookupEnv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if !ok {
		otelAgentAddr = "0.0.0.0:4317"
	}
	provider, closeFn := tracer.InitProvider(otelAgentAddr)

	srv := &Server{
		Tracer: provider,
		Router: mux.NewRouter(),
		stopCh: make(chan struct{}),
		tracerStopFn: closeFn,
		db:     db,
	}

	srv.routesV1()

	return srv
}

type Server struct {
	Tracer       *sdktrace.TracerProvider
	Router       *mux.Router
	stopCh       chan struct{}
	tracerStopFn func()
	db           *gorm.DB
}

// Run ...
func (s *Server) Run(ctx context.Context, port int) error {

	httpS := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: s.cors().Handler(s.requestID(s.Router)),
	}

	log.Info().Msgf("gopay-sh serving on port %d ", port)

	go func() {
		if err := httpS.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msgf("listen:%+s\n", err)
		}
	}()

	<-ctx.Done()

	log.Printf("server stopped")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

	s.tracerStopFn()

	return err

}

// checkServeErr checks the error from a .Serve() call to decide if it was a graceful shutdown
func (s *Server) checkServeErr(name string, err error) {
	if err != nil {
		if s.stopCh == nil {
			// a nil stopCh indicates a graceful shutdown
			log.Info().Msgf("graceful shutdown %s: %v", name, err)
		} else {
			log.Fatal().Msgf("%s: %v", name, err)
		}
	} else {
		log.Info().Msgf("graceful shutdown %s", name)
	}
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

func (s *Server) routesV1() {
	// healthcheck
	s.Router.HandleFunc("/", s.healthcheckHandler)
	s.Router.HandleFunc("/healthz", s.healthcheckHandler)
	s.Router.HandleFunc("/readyz", s.readinessHandler)

	// serve api
	api := s.Router.PathPrefix("/api/v1/").Subrouter()
	api.Use(
		// otelmux this is specific for otlp tracer
		otelmux.Middleware("booking.com", otelmux.WithTracerProvider(s.Tracer)),
	)

	meter := global.Meter("demo-server-meter")
	serverAttribute := attribute.String("server-attribute", "foo")
	commonLabels := []attribute.KeyValue{serverAttribute}
	requestCount, _ := meter.SyncInt64().Counter(
		"demo_server/request_counts",
		instrument.WithDescription("The number of requests received"),
	)

	api.Handle("/testy", otelhttp.NewHandler(http.HandlerFunc(s.healthcheckHandler), "/testy"))
	api.Handle("/testx", otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, req * http.Request) {
		requestCount.Add(req.Context(), 1, commonLabels...)

		labeler, _ := otelhttp.LabelerFromContext(req.Context())
		labeler.Add(attribute.Int("status_code", http.StatusOK))
	}), "/testx"))
}

func (s *Server) otel(h http.Handler) http.Handler {
	return otelhttp.NewHandler(h, "/ssss")
}

func (s *Server) requestID(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Request-Id") == "" {
			r.Header.Set("X-Request-Id", uuid.New().String())
		}

		log := log.With().
			Str("request_id", r.Header.Get("X-Request-Id")).
			Logger()

		ctx := log.WithContext(r.Context())
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) hc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("im alive"))
	}
}

func (s *Server) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("im alive"))
}

func (s *Server) readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("im ready to face the world"))
}

func gormDB(ctx context.Context) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s DB.name=%s password=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PASSWORD"))

	db, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn}), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("unable to open db connection")
	}

	err = db.Use(otelgorm.NewPlugin(otelgorm.WithDBName(os.Getenv("DB_NAME"))))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to set gorm plugin for opentelemetry ")
	}

	err = db.AutoMigrate(&booking.Show{})
	return db, err
}
