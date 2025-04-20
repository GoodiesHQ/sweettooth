package responses

import (
	"encoding/json"
	"net/http"
)

func JsonResponse(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
		}
	}
}
