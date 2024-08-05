package server

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// very simple middleware implementation
type Middleware func(http.Handler) http.Handler

// contain a list of the middleware functions which should be applied
type MiddlewareGroup struct {
	middleware []Middleware
}

// adds a function to the list of middleware to apply
func (mwg *MiddlewareGroup) Add(middleware ...Middleware) *MiddlewareGroup {
	for _, mw := range middleware {
		mwg.middleware = append(mwg.middleware, mw)
	}
	return mwg // allow chaining
}

// apply each of the middleware functions to the incoming request
func (mwg *MiddlewareGroup) Apply(handler http.Handler) http.Handler {
	for i := len(mwg.middleware) - 1; i >= 0; i-- {
		handler = mwg.middleware[i](handler)
	}
	return handler
}

func (mwg *MiddlewareGroup) ApplyFunc(handlerFunc func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return mwg.Apply(http.HandlerFunc(handlerFunc))
}

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
			evt = evt.Str("nodeid", *nodeid.(*string))
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
			if err, ok := err.(error); ok {
				evt = evt.Err(err)
			} else {
				evt = evt.Any("error", err)
			}
		}

		evt.Send()
	})
}

// helper type to log the status code when a request is complete
type statusResponseWriter struct {
	statusCode int
	w          http.ResponseWriter
}

func (w *statusResponseWriter) Write(data []byte) (int, error) {
	return w.w.Write(data)
}

func (w *statusResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *statusResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode // save the status code before writing
	w.w.WriteHeader(statusCode)
}

func MiddlewareJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// set JSON response header
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func MiddlewarePanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			var err error = nil

			if val := recover(); val != nil {
				switch val := val.(type) {
				case string:
					err = fmt.Errorf(val)
				case error:
					err = val
				default:
					err = fmt.Errorf("unknown panic: %v", val)
				}
				log.Error().Err(err).Send()
				JsonErr(w, r, http.StatusInternalServerError, errors.New("internal server error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
