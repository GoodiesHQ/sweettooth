package server

import (
	"net/http"
)

// very simple middleware implementation
type Middleware func(http.Handler) http.Handler

// contain a list of the middleware functions which should be applied
type MiddlewareGroup struct {
	middleware []Middleware
}

// adds a function to the list of middleware to apply
func (mwg *MiddlewareGroup) Add(middleware ...Middleware) *MiddlewareGroup {
	mwg.middleware = append(mwg.middleware, middleware...)
	return mwg
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
