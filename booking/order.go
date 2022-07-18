package booking

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTicketIsNotAvailable = fmt.Errorf("ticket is not available")
)

type OrderStatus string

const (
	Created  OrderStatus = "created"
	Reserved OrderStatus = "reserved"
	Rejected OrderStatus = "rejected"
)

type Order struct {
	gorm.Model
	ID     uuid.UUID `gorm:"type:uuid;not null;primaryKey"`
	ShowID uuid.UUID `gorm:"type:uuid;not null"`
	Status OrderStatus
}
