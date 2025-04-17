package responses

import (
	"encoding/json"
	"net/http"

	"github.com/goodieshq/sweettooth/internal/util"
)

func JsonResponse(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			util.SetRequestError(r, err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}
	}
}
