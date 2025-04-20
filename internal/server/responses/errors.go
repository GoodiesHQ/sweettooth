package responses

import (
	"errors"
	"net/http"

	"github.com/goodieshq/sweettooth/internal/server/requests"
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
		requests.SetRequestError(r, err)
		JsonErr(w, r, status, defaultErr)
	}
}

// JSON Errors
var ErrRegistrationTokenInvalid = CreateJsonErr(http.StatusUnauthorized, "the registration token is not found or is expired")
var ErrNodeTokenInvalid = CreateJsonErr(http.StatusUnauthorized, "the token is invalid or exired")
var ErrNodeUnauthorized = CreateJsonErr(http.StatusUnauthorized, "the token is not authorized")
var ErrNodeNotApproved = CreateJsonErr(http.StatusForbidden, "the node is not approved")
var ErrNodeNotFound = CreateJsonErr(http.StatusNotFound, "the node ID is not found")
var ErrOrgNotFound = CreateJsonErr(http.StatusNotFound, "the organization is not found")
var ErrInvalidPagination = CreateJsonErr(http.StatusBadRequest, "invalid pagination parameters")
var ErrInvalidRequestBody = CreateJsonErr(http.StatusBadRequest, "invalid request payload")
var ErrServiceUnavailable = CreateJsonErr(http.StatusServiceUnavailable, "service unavailable")
var ErrForbidden = CreateJsonErr(http.StatusForbidden, "insufficient privileges")
var ErrFormFailure = CreateJsonErr(http.StatusUnauthorized, "form submission failed")
var ErrLoginError = CreateJsonErr(http.StatusUnauthorized, "login failed")
var ErrServerError = CreateJsonErr(http.StatusInternalServerError, "internal server error")
var ErrInvalidOrgID = CreateJsonErr(http.StatusUnprocessableEntity, "the organization ID provided is invalid")
var ErrInvalidOrgRoles = CreateJsonErr(http.StatusUnprocessableEntity, "the organization roles provided are invalid")
var ErrInvalidNodeID = CreateJsonErr(http.StatusUnprocessableEntity, "the node ID provided is invalid")
var ErrInvalidJobID = CreateJsonErr(http.StatusUnprocessableEntity, "the job ID provided is invalid")
var ErrJobMissingOrExpired = CreateJsonErr(http.StatusNotFound, "this job ID is missing, expired, deleted, or has reached the attempt limit.")
var ErrJobAlreadyCompleted = CreateJsonErr(http.StatusConflict, "this job ID has already been completed")
var ErrDatabaseError = CreateJsonErr(http.StatusInternalServerError, "failed to connect to the database")
