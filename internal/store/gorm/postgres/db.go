package postgres

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	// Name is used for database name
	Name string
}

func (d Config) String() string {
	return fmt.Sprintf("host=%s port=%s user=%s DB.name=%s password=%s sslmode=disable", d.Host, d.Port, d.User, d.Name, d.Password)
}

func NewDB(config Config) *gorm.DB {
	dsn := config.String()

	db, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn}), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("unable to open db connection")
	}

	err = db.Use(otelgorm.NewPlugin(otelgorm.WithDBName(config.Name)))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to set gorm plugin for opentelemetry ")
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get sql db")
	}

	// Hardcode the max open connection for now
	sqlDB.SetMaxOpenConns(200)
	return db
}
