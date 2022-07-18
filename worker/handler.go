package worker

import (
	"context"
	"encoding/json"
	"os"

	"github.com/hibiken/asynq"

	"github.com/imrenagi/concurrent-booking/booking"
	"github.com/imrenagi/concurrent-booking/booking/services"
	"github.com/imrenagi/concurrent-booking/booking/stores"
	gormpg "github.com/imrenagi/concurrent-booking/internal/store/gorm/postgres"
)

type bookingService interface {
	Reserve(ctx context.Context, req services.ReservationRequest) (*booking.Ticket, error)
}

func newHandler() *handler {
	db := gormpg.NewDB(gormpg.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
	})

	bookingService := services.Booking{
		BookingRepository: stores.NewOrder(db),
		ShowRepository:    stores.NewShow(db),
	}

	return &handler{
		bookingService: bookingService,
	}
}

type handler struct {
	bookingService bookingService
}

func (h handler) HandleReservation(ctx context.Context, t *asynq.Task) error {
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
