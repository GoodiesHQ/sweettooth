package middleware

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func evtFromStatus(status int) *zerolog.Event {
	switch status / 100 {
	case 0:
		return log.Error()
	case 2:
		return log.Info()
	case 4:
		return log.Warn()
	case 5:
		return log.Error()
	default:
		return log.Info()
	}
}

func logNodeID(r *http.Request, evt **zerolog.Event) {
	// check if the node ID is in the request context (authorized node request)
	nodeid := r.Context().Value("nodeid")
	if nodeid != nil {
		if nodeid, ok := nodeid.(*uuid.UUID); ok {
			*evt = (*evt).Str("nodeid", nodeid.String())
		}
	}
}

// simple logging middleware to log requests as they come in
func MiddlewareLogger() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// take a before and after timestamp to determine how long the request takes
			rw := logResponseWriter{w: w} // log the status code from future handlers
			t1 := time.Now()
			next.ServeHTTP(&rw, r)
			t2 := time.Now()

			evt := evtFromStatus(rw.statusCode)

			// add basic request information like the method and path
			evt = evt.Str("method", r.Method).Str("path", r.URL.Path)

			logNodeID(r, &evt)

			// add the latency in MS
			evt = evt.Int64("latency_ms", t2.Sub(t1).Milliseconds())

			// add the response status
			if rw.statusCode != 0 {
				evt = evt.Int("status_code", rw.statusCode).Str("status", http.StatusText(rw.statusCode))
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
}

// helper type to log the status code when a request is complete
type logResponseWriter struct {
	statusCode int
	w          http.ResponseWriter
}

func (w *logResponseWriter) Write(data []byte) (int, error) {
	return w.w.Write(data)
}

func (w *logResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *logResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode // save the status code before writing
	w.w.WriteHeader(statusCode)
}
