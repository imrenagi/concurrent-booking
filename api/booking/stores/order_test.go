package stores_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/imrenagi/concurrent-booking/api/booking"
	"github.com/imrenagi/concurrent-booking/api/booking/stores"
)

func postgresC() (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:13-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "booking",
			"POSTGRES_USER":     "booking",
			"POSTGRES_PASSWORD": "booking",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithPollInterval(1 * time.Second),
	}

	return testcontainers.GenericContainer(context.TODO(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

func gormFromContainer(ctx context.Context, container testcontainers.Container) (*gorm.DB, error) {
	port, err := container.Ports(ctx)
	log.Debug().Msgf("%v %v", port["5432/tcp"][0].HostIP, port["5432/tcp"][0].HostPort)

	dsn := fmt.Sprintf("host=%s port=%s user=%s DB.name=%s password=%s sslmode=disable",
		port["5432/tcp"][0].HostIP,
		port["5432/tcp"][0].HostPort,
		"booking",
		"booking",
		"booking")

	db, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn}), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("unable to open db connection")
	}

	err = db.AutoMigrate(&booking.Show{})
	return db, err
}

func TestConcurrentBooking(t *testing.T) {
	ctx := context.TODO()
	postgresC, err := postgresC()
	assert.NoError(t, err)
	defer postgresC.Terminate(ctx)

	db, err := gormFromContainer(ctx, postgresC)
	assert.NoError(t, err)

	repo := stores.NewShow(db)
	id := uuid.New()
	err = repo.Save(context.TODO(), &booking.Show{
		ID:               id,
		RemainingTickets: 10,
	})
	assert.NoError(t, err)

	ticketResultChan := make(chan booking.Ticket, 10)

	orderRepo := stores.NewOrder(db)

	for i := 0; i < 10; i++ {
		go func(i int) {
			err := orderRepo.Reserve(context.TODO(), id)
			if err != nil {
				log.Fatal().Msg("error")
			}
			ticketResultChan <- booking.Ticket{}
		}(i)
	}

	doneChan := make(chan bool)
	go func() {
		<-time.After(3 * time.Second)
		doneChan <- true
	}()

	ticketCount := 0
jump:
	for {
		select {
		case <-ticketResultChan:
			ticketCount++
		case <-doneChan:
			break jump
		}
	}
	assert.Equal(t, 10, ticketCount)
	show, err := repo.FindConcertByID(context.TODO(), id)
	assert.NoError(t, err)
	assert.Equal(t, 0, show.RemainingTickets)
}
