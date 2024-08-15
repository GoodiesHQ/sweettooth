package client

import (
	"errors"
	"net/http"

	"github.com/goodieshq/sweettooth/pkg/api"
	"github.com/rs/zerolog/log"
)

var (
	ErrNodeNotApproved       = errors.New("node is not approved")
	ErrNodeNotRegistered     = errors.New("node is not registered")
	ErrNodeAlreadyRegistered = errors.New("node is already registered")
)

type StatusMap map[int]error

// status map that essentially always applies. some endpoints may differ
var GeneralStatusMap = StatusMap{
	http.StatusNoContent: nil,                  // always a successful response
	http.StatusOK:        nil,                  // always a successful response
	http.StatusForbidden: ErrNodeNotApproved,   // always an unapproved node
	http.StatusNotFound:  ErrNodeNotRegistered, // always an unregistered node
}

func LogApiErr(apierr api.ErrorResponse) {
	log.Error().
		Int("status_code", apierr.StatusCode).
		Str("status", apierr.Status).
		Msg(apierr.Message)
}

func LogIfApiErr(err error) bool {
	var apierr api.ErrorResponse
	if errors.As(err, &apierr) {
		LogApiErr(apierr)
		return true
	}
	return false
}
