package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type BookingRepository interface {
	Booking(ctx context.Context, ID uuid.UUID) error
}

type Handler struct {
	BookingRepository BookingRepository
}

type BookingRequest struct {
	ShowID uuid.UUID `json:"show_id"`
}

func (h Handler) Booking() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var body BookingRequest
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			log.Debug().Err(err).Msg("unable to parse the update addon request")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = h.BookingRepository.Booking(r.Context(), body.ShowID)
		if err != nil {
			log.Debug().Err(err).Msg("failed process booking")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}
