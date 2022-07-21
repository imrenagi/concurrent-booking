package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/imrenagi/concurrent-booking/booking"
)

var tracer = otel.Tracer("github.com/imrenagi/concurrent-booking/booking/services")

type BookingRepository interface {
	FindOrderByID(ctx context.Context, ID uuid.UUID) (*booking.Order, error)
	Reserve(ctx context.Context, ID uuid.UUID) error
	Save(ctx context.Context, order *booking.Order) error
	Create(ctx context.Context, order *booking.Order) error
}

type ShowRepository interface {
	FindConcertByID(ctx context.Context, ID uuid.UUID) (*booking.Show, error)
}

type Booking struct {
	BookingRepository BookingRepository
	ShowRepository    ShowRepository
	Dispatcher        *asynq.Client
}

type BookingRequest struct {
	ShowID uuid.UUID `json:"show_id"`
}

func (b Booking) Book(ctx context.Context, req BookingRequest) (*booking.Ticket, error) {
	ctx, parentSpan := tracer.Start(ctx, "booking.BookV1")
	defer parentSpan.End()

	err := b.BookingRepository.Reserve(ctx, req.ShowID)
	if err != nil {
		return nil, err
	}
	return &booking.Ticket{}, nil
}

func (b Booking) BookV2(ctx context.Context, req BookingRequest) (*booking.Order, error) {
	ctx, parentSpan := tracer.Start(ctx, "booking.BookV2")
	defer parentSpan.End()

	parentSpan.AddEvent("creating new order id")
	order := booking.Order{
		ID:     uuid.New(),
		ShowID: req.ShowID,
		Status: booking.Created,
	}

	err := b.BookingRepository.Create(ctx, &order)
	if err != nil {
		parentSpan.RecordError(err)
		return nil, err
	}

	parentSpan.AddEvent("creating asynq task")
	task, err := NewReservationTask(ctx, order)
	if err != nil {
		parentSpan.RecordError(err)
		return nil, err
	}

	parentSpan.AddEvent("Adding task to queue")
	taskInfo, err := b.Dispatcher.EnqueueContext(ctx, task, asynq.Queue("critical"))
	if err != nil {
		parentSpan.RecordError(err)
		return nil, err
	}
	parentSpan.AddEvent("task is created", trace.WithAttributes(attribute.String("task_info_id", taskInfo.ID)))

	return &order, nil
}

type ReservationRequest struct {
	ShowID  uuid.UUID `json:"show_id"`
	OrderID uuid.UUID `json:"order_id"`
}

func (b Booking) Reserve(ctx context.Context, req ReservationRequest) (*booking.Ticket, error) {
	ctx, parentSpan := tracer.Start(ctx, "booking.Reserve")
	defer parentSpan.End()
	err := b.BookingRepository.Reserve(ctx, req.ShowID)
	if err != nil && err != booking.ErrTicketIsNotAvailable {
		return nil, err
	}

	if err == booking.ErrTicketIsNotAvailable {
		log.Debug().Msgf("order is rejected")
		err = b.setOrderToRejected(ctx, req.OrderID)
		if err != nil {
			return nil, err
		}
	} else {
		err = b.setOrderToReserved(ctx, req.OrderID)
		if err != nil {
			return nil, err
		}
	}

	return &booking.Ticket{}, nil
}

func (b Booking) setOrderToReserved(ctx context.Context, id uuid.UUID) error {
	ctx, parentSpan := tracer.Start(ctx, "booking.setOrderToReserved")
	defer parentSpan.End()
	order, err := b.BookingRepository.FindOrderByID(ctx, id)
	if err != nil {
		return err
	}
	order.Status = booking.Reserved

	err = b.BookingRepository.Save(ctx, order)
	if err != nil {
		return err
	}
	return nil
}

func (b Booking) setOrderToRejected(ctx context.Context, id uuid.UUID) error {
	ctx, parentSpan := tracer.Start(ctx, "booking.setOrderToRejected")
	defer parentSpan.End()
	order, err := b.BookingRepository.FindOrderByID(ctx, id)
	if err != nil {
		return err
	}
	order.Status = booking.Rejected

	err = b.BookingRepository.Save(ctx, order)
	if err != nil {
		return err
	}
	return nil
}
