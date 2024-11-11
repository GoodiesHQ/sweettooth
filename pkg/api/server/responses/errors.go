package responses

import (
	"errors"
	"net/http"

	"github.com/goodieshq/sweettooth/internal/util"
	"github.com/goodieshq/sweettooth/pkg/api"
)

func JsonErr(w http.ResponseWriter, r *http.Request, status int, err error) {
	JsonResponse(w, r, status, &api.ErrorResponse{Status: "error", Message: err.Error()})
}

// the default messages are always sent as a response, but custom errors will show up in the logs instead
func CreateJsonErr(status int, defaultMessage string) func(http.ResponseWriter, *http.Request, error) {
	// create an error with the default message
	defaultErr := errors.New(defaultMessage)

	// the default handler will log the error your provide, but send a default error to the end user
	return func(w http.ResponseWriter, r *http.Request, err error) {
		if err == nil {
			err = defaultErr
		}
		util.SetRequestError(r, err)
		JsonErr(w, r, status, defaultErr)
	}
}

// JSON Errors
var ErrJsonRegistrationTokenInvalid = CreateJsonErr(http.StatusUnauthorized, "the registration token is not found or is expired")
var ErrJsonNodeTokenInvalid = CreateJsonErr(http.StatusUnauthorized, "the token is invalid or exired")
var ErrJsonNodeUnauthorized = CreateJsonErr(http.StatusUnauthorized, "the token is not authorized")
var ErrJsonNodeNotApproved = CreateJsonErr(http.StatusForbidden, "the node is not approved")
var ErrJsonNodeNotFound = CreateJsonErr(http.StatusNotFound, "the node ID is not found")
var ErrJsonOrgNotFound = CreateJsonErr(http.StatusNotFound, "the organization is not found")
var ErrJsonInvalidRequestBody = CreateJsonErr(http.StatusBadRequest, "invalid request payload")
var ErrJsonServiceUnavailable = CreateJsonErr(http.StatusServiceUnavailable, "service unavailable")
var ErrJsonServerError = CreateJsonErr(http.StatusInternalServerError, "internal server error")
var ErrJsonInvalidJobID = CreateJsonErr(http.StatusUnprocessableEntity, "the job ID provided is invalid")
var ErrJsonJobMissingOrExpired = CreateJsonErr(http.StatusNotFound, "this job ID is missing, expired, deleted, or has reached the attempt limit.")
var ErrJsonJobAlreadyCompleted = CreateJsonErr(http.StatusConflict, "this job ID has already been completed")
