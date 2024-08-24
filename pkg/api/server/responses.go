package server

import (
	"encoding/json"
	"net/http"

	"github.com/goodieshq/sweettooth/internal/util"
	"github.com/goodieshq/sweettooth/pkg/api"
)

func JsonErr(w http.ResponseWriter, r *http.Request, status int, err error) {
	JsonResponse(w, r, status, &api.ErrorResponse{Status: "error", Message: err.Error()})
}

func JsonResponse(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			util.SetRequestError(r, err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}
	}
}
