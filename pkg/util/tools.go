package util

import (
	"context"
	"net/http"
)

func SetRequestContext(r *http.Request, ctx context.Context) {
	*r = *r.WithContext(ctx)
}

func SetRequestError(r *http.Request, err error) {
	SetRequestContext(r, context.WithValue(r.Context(), "error", err))
}

func SetRequestNodeID(r *http.Request, nodeid string) {
	SetRequestContext(r, context.WithValue(r.Context(), "nodeid", &nodeid))
}
