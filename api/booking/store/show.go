package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/imrenagi/concurrent-booking/api/booking"
)

func NewShow(db *gorm.DB) *Show {
	return &Show{
		db: db,
	}
}

type Show struct {
	db *gorm.DB
}

func (c Show) Booking(ctx context.Context, ID uuid.UUID) error {
	return c.db.Transaction(func(tx *gorm.DB) error {
		var show *booking.Show
		err := tx.WithContext(ctx).
			Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).
			Where("id = ?", ID).
			First(&show).Error
		if err != nil {
			return err
		}

		if show.RemainingTickets-1 < 0 {
			return fmt.Errorf("ticket is not available")
		}
		show.RemainingTickets -= 1

		log.Debug().Int("remaining_tickets", show.RemainingTickets).Msg("remaining tickets have been decreased")

		return tx.Save(&show).Error
	})
}

func (c Show) FindConcertByID(ctx context.Context, ID uuid.UUID) (*booking.Show, error) {
	var show *booking.Show
	err := c.db.WithContext(ctx).
		Where("id = ?", ID).
		First(&show).Error
	return show, err
}

func (c Show) Save(ctx context.Context, concert *booking.Show) error {
	return c.db.WithContext(ctx).Save(&concert).Error
}
