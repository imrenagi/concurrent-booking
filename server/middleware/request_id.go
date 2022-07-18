package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func RequestID(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Request-Id") == "" {
			r.Header.Set("X-Request-Id", uuid.New().String())
		}

		log := log.With().
			Str("request_id", r.Header.Get("X-Request-Id")).
			Logger()

		ctx := log.WithContext(r.Context())
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
