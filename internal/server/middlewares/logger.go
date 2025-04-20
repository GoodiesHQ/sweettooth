package middlewares

import (
	"net/http"
	"time"

	"github.com/goodieshq/sweettooth/internal/server/requests"
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

// simple logging middleware to log requests as they come in
func MiddlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// take a before and after timestamp to determine how long the request takes
		rw := logResponseWriter{w: w} // log the status code from future handlers
		t1 := time.Now()
		next.ServeHTTP(&rw, r)
		t2 := time.Now()

		evt := evtFromStatus(rw.statusCode)

		// add basic request information like the method and path
		evt = evt.Str("method", r.Method).Str("path", r.URL.Path)

		if state := requests.State(r); state != nil && state.IsNodeRequest() {
			evt = evt.Str("nodeid", state.NodeID.String())
		}

		// add the latency in MS
		evt = evt.Int64("latency_ms", t2.Sub(t1).Milliseconds())

		// add the response status
		evt = evt.Int("status_code", rw.statusCode).Str("status", http.StatusText(rw.statusCode))

		// check if there is an error
		err := requests.Err(r)
		if err != nil {
			evt = evt.Err(err)
		}

		evt.Send()
	})
}

// helper type to log the status code when a request is complete
type logResponseWriter struct {
	statusCode int
	w          http.ResponseWriter
}

func (w *logResponseWriter) Write(data []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	return w.w.Write(data)
}

func (w *logResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *logResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode // save the status code before writing
	w.w.WriteHeader(statusCode)
}
