package server

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func JsonErr(w http.ResponseWriter, r *http.Request, status int, err error) {
	JsonResponse(w, status, &ErrorResponse{Status: "error", Message: err.Error()})
}

func JsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
