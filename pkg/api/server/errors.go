package server

import (
	"errors"
	"net/http"

	"github.com/goodieshq/sweettooth/pkg/util"
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

var ErrNodeTokenInvalid = CreateErr(http.StatusUnauthorized, "the token is invalid or exired")
var ErrNodeUnauthorized = CreateErr(http.StatusUnauthorized, "the token is not authorized")
var ErrNodeNotApproved = CreateErr(http.StatusUnauthorized, "the node is not approved")
var ErrNodeNotFound = CreateErr(http.StatusNotFound, "the node ID is not found")
var ErrInvalidRequestBody = CreateErr(http.StatusBadRequest, "invalid request body")
