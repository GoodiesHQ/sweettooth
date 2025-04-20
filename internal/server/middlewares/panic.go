package middlewares

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/goodieshq/sweettooth/internal/server/responses"
	"github.com/rs/zerolog/log"
)

// implement middleware to prevent any kind of panic
func MiddlewarePanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			var err error = nil

			// if there is ever a panic...
			if val := recover(); val != nil {
				switch val := val.(type) {
				case string:
					err = errors.New(val)
				case error:
					err = val
				default:
					err = fmt.Errorf("unknown panic: %v", val)
				}
				// ... log the panic value as an error and return a 500
				log.Warn().Err(err).Msg("panic recovered")
				responses.ErrServerError(w, r, err)
				return
			}
		}()
		next.ServeHTTP(w, r)
	})
}
