package stores

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

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
