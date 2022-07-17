package server

import (
	"os"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/imrenagi/concurrent-booking/api/booking"
	gormpg "github.com/imrenagi/concurrent-booking/api/internal/store/gorm/postgres"
)

func db() *gorm.DB {
	db := gormpg.NewDB(gormpg.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
	})
	if err := db.AutoMigrate(&booking.Show{}, &booking.Order{}); err != nil {
		log.Fatal().Err(err).Msgf("database migration failed")
	}
	return db
}
