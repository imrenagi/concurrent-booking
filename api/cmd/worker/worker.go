package main

import (
	"context"
	"encoding/json"
	"fmt"
	// "log"
	"os"

	// "log"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/imrenagi/concurrent-booking/api/booking"
	"github.com/imrenagi/concurrent-booking/api/booking/services"
	"github.com/imrenagi/concurrent-booking/api/booking/stores"
	"github.com/imrenagi/concurrent-booking/api/pkg/tracer"
)

const redisAddr = "127.0.0.1:6379"

var trc = otel.Tracer("github.com/imrenagi/concurrent-booking/api/cmd/worker")

func main() {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 20,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)
	otelAgentAddr, ok := os.LookupEnv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if !ok {
		otelAgentAddr = "0.0.0.0:4317"
	}
	_, closeFn := tracer.InitProvider("reservation-worker", otelAgentAddr)
	defer closeFn()

	handler := NewHandler()

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.Use(tracingMiddleware)
	mux.HandleFunc("booking:reserve", handler.HandleReservation)

	if err := srv.Run(mux); err != nil {
		log.Fatal().Msgf("could not run server: %v", err)
	}
}

func tracingMiddleware(next asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		var tp services.TaskPayload
		err := json.Unmarshal(t.Payload(), &tp)
		if err != nil {
			return err
		}
		spanCtx, _ := tracer.ConstructNewSpanContext(tracer.NewRequest{
			TraceID: tp.Context.TraceID,
			SpanID:  tp.Context.SpanID,
		})
		ctx = trace.ContextWithSpanContext(ctx, spanCtx)
		payload, err := json.Marshal(tp.Data)
		if err != nil {
			return err
		}
		return next.ProcessTask(ctx, asynq.NewTask(t.Type(), payload))
	})
}

func NewHandler() *Handler {
	ctx := context.Background()
	db, err := gormDB(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to start db")
	}

	bookingService := services.Booking{
		BookingRepository: stores.NewOrder(db),
		ShowRepository:    stores.NewShow(db),
	}

	return &Handler{
		bookingService: bookingService,
	}
}

type BookingService interface {
	Reserve(ctx context.Context, req services.ReservationRequest) (*booking.Ticket, error)
}

type Handler struct {
	bookingService BookingService
}

func (h Handler) HandleReservation(ctx context.Context, t *asynq.Task) error {
	ctx, span := trc.Start(ctx, "handler.HandleReservation")
	defer span.End()

	var order booking.Order
	err := json.Unmarshal(t.Payload(), &order)
	if err != nil {
		return err
	}

	_, err = h.bookingService.Reserve(ctx, services.ReservationRequest{
		ShowID:  order.ShowID,
		OrderID: order.ID,
	})
	if err != nil && err != booking.ErrTicketIsNotAvailable {
		return err
	}

	return nil
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

	sqlDB, err := db.DB()
	sqlDB.SetMaxOpenConns(200)

	return db, err
}
