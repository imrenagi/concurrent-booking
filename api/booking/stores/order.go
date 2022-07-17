package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/imrenagi/concurrent-booking/api/booking"
)

func NewOrder(db *gorm.DB) *Order {
	return &Order{db: db}
}

type Order struct {
	db *gorm.DB
}

func (o Order) FindOrderByID(ctx context.Context, ID uuid.UUID) (*booking.Order, error) {
	var order *booking.Order
	err := o.db.WithContext(ctx).
		Where("id = ?", ID).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (o Order) Save(ctx context.Context, order *booking.Order) error {
	return o.db.WithContext(ctx).Save(&order).Error
}

func (o Order) Create(ctx context.Context, order *booking.Order) error {
	return o.db.WithContext(ctx).Create(&order).Error
}

func (o Order) Reserve(ctx context.Context, showID uuid.UUID) error {
	return o.db.Transaction(func(tx *gorm.DB) error {
		var show *booking.Show
		err := tx.WithContext(ctx).
			Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).
			Where("id = ?", showID).
			First(&show).Error
		if err != nil {
			return err
		}

		if show.RemainingTickets-1 < 0 {
			return booking.ErrTicketIsNotAvailable
		}
		show.RemainingTickets -= 1

		log.Debug().Int("remaining_tickets", show.RemainingTickets).Msg("remaining tickets have been decreased")

		return tx.WithContext(ctx).Save(&show).Error
	})
}
