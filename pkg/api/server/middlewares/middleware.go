package middlewares

import (
	"net/http"
)

var all = []Middleware{
	MiddlewareJSON(),
	MiddlewareLogger(),
	MiddlewarePanic(),
}

type Middleware func(next http.Handler) http.Handler
type MiddlewareGroup []Middleware

func (mwg *MiddlewareGroup) Add(middleware ...Middleware) *MiddlewareGroup {
	*mwg = append(*mwg, middleware...)
	return mwg
}

// apply each of the middleware functions to the incoming request
func (mwg *MiddlewareGroup) Apply(handler http.Handler) http.Handler {
	for i := len(*mwg) - 1; i >= 0; i-- {
		handler = (*mwg)[i](handler)
	}
	return handler
}

func (mwg *MiddlewareGroup) ApplyFunc(handlerFunc func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return mwg.Apply(http.HandlerFunc(handlerFunc))
}
