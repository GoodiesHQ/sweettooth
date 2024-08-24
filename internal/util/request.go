package util

import (
	"context"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	DEFAULT_ATTEMPTS_MAX = 5   //
	LIMIT_ATTEMPTS_MAX   = 100 // set a maximum limit of 100 attempts
)

// get the query parameter of attempts max
func RequestQueryAttemptsMax(r *http.Request) int {
	var attemptsMax int = 0

	// get and parse the attempts_max parameter
	if attemptsMaxStr := r.URL.Query().Get("attempts_max"); attemptsMaxStr != "" {
		// attempts_max should be an integer
		i, err := strconv.Atoi(attemptsMaxStr)
		if err != nil {
			log.Warn().Str("attempts_max", attemptsMaxStr).Msg("invalid attempts_max value")
		} else {
			if i > LIMIT_ATTEMPTS_MAX {
				i = LIMIT_ATTEMPTS_MAX
			}
			attemptsMax = i
		}
	}

	if attemptsMax == 0 {
		// default or invalid attepmts_max value set to a sane default
		attemptsMax = DEFAULT_ATTEMPTS_MAX
	} else if attemptsMax < 0 {
		// if any negative value was provided, set it to -1
		attemptsMax = -1
	}

	return attemptsMax
}

func Rid(r *http.Request) *uuid.UUID {
	return r.Context().Value("nodeid").(*uuid.UUID)
}

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
