package apiweb

import (
	"net/http"

	"github.com/goodieshq/sweettooth/internal/server/responses"
)

// POST /api/v1/web/login
func (h *ApiWebHandler) HandlePostWebLogin(w http.ResponseWriter, r *http.Request) {
	// TODO: process login
	responses.JsonResponse(w, r, http.StatusOK, nil)
}