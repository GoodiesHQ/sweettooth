package middlewares

import (
	"net/http"

	"github.com/goodieshq/sweettooth/internal/crypto"
	"github.com/goodieshq/sweettooth/internal/server/cache"
	"github.com/goodieshq/sweettooth/internal/server/core"
	"github.com/goodieshq/sweettooth/internal/server/requests"
	"github.com/goodieshq/sweettooth/internal/server/responses"
	"github.com/rs/zerolog/log"
)

const (
	DEV_BYPASS_WEBAUTH = true
)

func MiddlewareAuthWeb(core core.Core, cache cache.Cache, secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if DEV_BYPASS_WEBAUTH {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Trace().Msg("starting middleware web auth")
			// extract the bearer token from the Authorization header
			tokenString := ExtractBearerToken(r)
			if tokenString == "" {
				responses.ErrNodeTokenInvalid(w, r, nil)
				return
			}

			// verify that the JWT was signed with the key that it says it was signed with
			userid, superAdmin, orgRoles, err := crypto.VerifyWebJWT(tokenString, secret)
			if err != nil {
				log.Debug().Err(err).Msg("jwt was unverified")
				responses.ErrNodeTokenInvalid(w, r, err)
				return
			}

			// set the super admin flag (may be returned even upon error)
			r = requests.WithRequestSuperAdmin(r, superAdmin)

			// set the user ID (may be returned even upon error)
			r = requests.WithRequestUserID(r, userid)

			// set the org roles
			r = requests.WithRequestOrgRoles(r, orgRoles)

			next.ServeHTTP(w, r)
		})
	}
}
