package middlewares

import (
	"net/http"

	"github.com/goodieshq/sweettooth/internal/server/requests"
)

func MiddlewareState(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state := &requests.RequestState{}
		r = requests.WithRequestState(r, state)
		next.ServeHTTP(w, r)
	})
}