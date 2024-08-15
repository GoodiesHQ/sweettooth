package server

import "net/http"

func MiddlewareJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// set JSON response header
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
