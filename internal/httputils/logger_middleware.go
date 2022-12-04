package httputils

import (
	"net/http"

	"github.com/rs/zerolog"
)

func LoggerMiddleware(logger zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(logger.WithContext(r.Context())))
		})
	}
}
