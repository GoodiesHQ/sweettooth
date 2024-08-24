package server

import (
	"errors"
	"net/http"

	"github.com/goodieshq/sweettooth/internal/util"
)

// the default messages are always sent as a response, but custom errors will show up in the logs instead
func CreateErr(status int, defaultMessage string) func(http.ResponseWriter, *http.Request, error) {
	defaultErr := errors.New(defaultMessage)
	return func(w http.ResponseWriter, r *http.Request, err error) {
		if err == nil {
			err = defaultErr
		}
		util.SetRequestError(r, err)
		JsonErr(w, r, status, defaultErr)
	}
}

var ErrRegistrationTokenInvalid = CreateErr(http.StatusUnauthorized, "the registration token is not found or expired")
var ErrNodeTokenInvalid = CreateErr(http.StatusUnauthorized, "the token is invalid or exired")
var ErrNodeUnauthorized = CreateErr(http.StatusUnauthorized, "the token is not authorized")
var ErrNodeNotApproved = CreateErr(http.StatusForbidden, "the node is not approved")
var ErrNodeNotFound = CreateErr(http.StatusNotFound, "the node ID is not found")
var ErrOrgNotFound = CreateErr(http.StatusNotFound, "the organization is not found")
var ErrInvalidRequestBody = CreateErr(http.StatusBadRequest, "invalid request payload")
var ErrServiceUnavailable = CreateErr(http.StatusServiceUnavailable, "service unavailable")
var ErrServerError = CreateErr(http.StatusInternalServerError, "internal server error")
var ErrInvalidJobID = CreateErr(http.StatusUnprocessableEntity, "the job ID provided is invalid")
var ErrJobMissingOrExpired = CreateErr(http.StatusNotFound, "this job ID is missing, expired, deleted, or has reached the attempt limit.")
var ErrJobAlreadyCompleted = CreateErr(http.StatusConflict, "this job ID has already been completed")
