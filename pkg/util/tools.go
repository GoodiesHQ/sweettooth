package util

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

func SetRequestContext(r *http.Request, ctx context.Context) {
	*r = *r.WithContext(ctx)
}

func SetRequestContextValue(r *http.Request, key string, value interface{}) {
	SetRequestContext(r, context.WithValue(r.Context(), key, value))
}

func SetRequestError(r *http.Request, err error) {
	SetRequestContextValue(r, "error", err)
}

func SetRequestNodeID(r *http.Request, nodeid uuid.UUID) {
	SetRequestContextValue(r, "nodeid", &nodeid)
}

func Countdown(prefix string, n uint, suffix string) {
	for i := n; i > 0; i-- {
		fmt.Printf("%s%d%s", prefix, i, suffix)
		time.Sleep(time.Second * 1) // sleep for 1 second
		fmt.Print("\r\033[K")       // return to beginning of line and ascii escape to clear line
	}
}

func Dumps(obj any) string {
	data, _ := json.MarshalIndent(obj, "", "  ")
	return string(data)
}

func Rid(r *http.Request) *uuid.UUID {
	return r.Context().Value("nodeid").(*uuid.UUID)
}

func IsFile(path string) bool {
	if info, err := os.Stat(path); err != nil {
		return false
	} else {
		return !info.IsDir() // it should exist and NOT be a directory
	}
}
