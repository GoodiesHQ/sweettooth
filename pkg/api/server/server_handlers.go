package server

import (
	"encoding/json"
	"net/http"

	"github.com/goodieshq/sweettooth/pkg/api"
)

// GET /api/v1/node/check
func (srv *SweetToothServer) handleGetNodeCheck(w http.ResponseWriter, r *http.Request) {
	// a successful check should only yield a 204
	// this indicates a properly formed and valid JWT, node ID exists in the database and is enabled
	JsonResponse(w, http.StatusNoContent, nil)
}

// POST /api/v1/node/register
func (srv *SweetToothServer) handlePostNodeRegister(w http.ResponseWriter, r *http.Request) {
	var req api.RegistrationRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	// node registration
}
