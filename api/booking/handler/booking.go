package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/imrenagi/concurrent-booking/api/booking"
	"github.com/imrenagi/concurrent-booking/api/booking/services"
)

type BookingService interface {
	Book(ctx context.Context, req services.BookingRequest) (*booking.Ticket, error)
	BookV2(ctx context.Context, req services.BookingRequest) (*booking.Order, error)
}

type Handler struct {
	Service BookingService
}

func (h Handler) Booking() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var body services.BookingRequest
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			log.Debug().Err(err).Msg("unable to parse the update addon request")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = h.Service.Book(r.Context(), body)
		if err != nil {
			log.Debug().Err(err).Msg("failed process booking")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}

func (h Handler) BookingV2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var body services.BookingRequest
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			log.Debug().Err(err).Msg("unable to parse the update addon request")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		order, err := h.Service.BookV2(r.Context(), body)
		if err != nil {
			log.Debug().Err(err).Msg("failed process booking")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		b, _ := json.Marshal(order)
		w.Header().Add("Content-Type", "application/json")
		w.Write(b)
	}
}
