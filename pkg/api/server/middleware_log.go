package server

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// simple logging middleware to log requests as they come in
func MiddlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// take a before and after timestamp to determine how long the request takes
		sw := statusResponseWriter{w: w} // log the status code from future handlers
		t1 := time.Now()
		next.ServeHTTP(&sw, r)
		t2 := time.Now()

		var evt *zerolog.Event

		switch sw.statusCode / 100 {
		case 0:
			evt = log.Error()
		case 2:
			evt = log.Info()
		case 4:
			evt = log.Warn()
		case 5:
			evt = log.Error()
		default:
			evt = log.Info()
		}

		// add basic request information like the method and path
		evt = evt.Str("method", r.Method).Str("path", r.URL.Path)

		// check if the node ID is in the request context (authorized node request)
		nodeid := r.Context().Value("nodeid")
		if nodeid != nil {
			if nodeid, ok := nodeid.(*uuid.UUID); ok {
				evt = evt.Str("nodeid", nodeid.String())
			}
		}

		// add the latency in MS
		evt = evt.Int64("latency_ms", t2.Sub(t1).Milliseconds())

		// add the response status
		if sw.statusCode != 0 {
			evt = evt.Int("status_code", sw.statusCode).Str("status", http.StatusText(sw.statusCode))
		}

		// check if there is an error
		err := r.Context().Value("error")
		if err != nil {
			if errParsed, ok := err.(error); ok {
				evt = evt.Err(errParsed)
			} else {
				evt = evt.Any("error", err)
			}
		}

		evt.Send()
	})
}
