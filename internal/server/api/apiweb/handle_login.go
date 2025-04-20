package apiweb

import (
	"encoding/json"
	"net/http"

	"github.com/goodieshq/sweettooth/internal/crypto"
	"github.com/goodieshq/sweettooth/internal/server/responses"
	"github.com/goodieshq/sweettooth/internal/server/roles"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// POST /api/v1/web/login
func (h *ApiWebHandler) HandlePostWebLogin(secret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: process login properly
		var creds LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			responses.ErrFormFailure(w, r, err)
			return
		}

		log.Info().Str("username", creds.Username).Str("password", creds.Password).Msg("Login attempt")
		if creds.Username == "admin" && creds.Password == "admin123" {
			token, err := crypto.CreateWebJWT(
				uuid.MustParse("f3d670ad-7189-40be-b26e-ba4fbc9e2469"),
				true,
				roles.OrgRoles{},
				secret,
			)
			if err == nil {
				responses.JsonResponse(w, r, http.StatusOK, map[string]string{
					"message": "Login successful",
					"token":   token,
				})
				return
			}
		}
		responses.ErrLoginError(w, r, nil)
		return
	}
}
