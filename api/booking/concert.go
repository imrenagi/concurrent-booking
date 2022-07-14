package booking

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Show struct {
	gorm.Model
	ID               uuid.UUID `gorm:"type:uuid;not null;primaryKey"`
	RemainingTickets int
}
