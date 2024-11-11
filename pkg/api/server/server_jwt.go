package server

import (
	"errors"
	"net/http"
	"strings"

	"github.com/goodieshq/sweettooth/internal/crypto"
	"github.com/goodieshq/sweettooth/internal/util"
	"github.com/goodieshq/sweettooth/pkg/api"
	"github.com/goodieshq/sweettooth/pkg/api/server/responses"
	"github.com/rs/zerolog/log"
)

// Extract the bearer token from the Authorization header
func extractAuthToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// all valid tokens will be in the form of "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// sends an API error if there is an error to send. If err is nil, returns immediately as a noop
func nodeAuthErr(w http.ResponseWriter, r *http.Request, err error, node *api.Node) bool {
	if err == nil {
		if node == nil {
			responses.ErrJsonNodeNotFound(w, r, nil)
			return true
		}
		/* no errors errors and the node exists */
	} else {
		responses.ErrJsonServerError(w, r, err)
		return true
	}

	if !node.Approved {
		log.Debug().Msg("node is registered but not approved")
		responses.ErrJsonNodeNotApproved(w, r, nil)
		return true
	}

	responses.ErrJsonServerError(w, r, err)
	return true
}

// Middleware for handling endpoints which are exclusively used by nodes running sweettooth clients
func (srv *SweetToothServer) MiddlewareNodeAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Trace().Msg("starting middleware node auth")
		// extract the bearer token from the Authorization header
		tokenString := extractAuthToken(r)
		if tokenString == "" {
			responses.ErrJsonNodeTokenInvalid(w, r, nil)
			return
		}
		log.Trace().Msg("extracted bearer token")

		nodeid, _, err := crypto.VerifyNodeJWT(tokenString)
		nodeidString := nodeid.String()
		if nodeidString != "" {
			// set the node ID (may be returned even upon error)
			util.SetRequestNodeID(r, nodeid)
			log.Trace().Str("nodeid", nodeidString).Msg("node id added to the request")
		}
		if err != nil {
			log.Debug().Err(err).Msg("jwt was unverified")
			responses.ErrJsonNodeTokenInvalid(w, r, err)
			return
		}

		/* At this point, all we know is that the signature is valid and well-formed. Check cache/db for validity */
		// authorized, found := srv.cacheGetAuth(nodeidString)
		found, authorized := srv.cache.GetAuth(nodeidString)
		if !found {
			log.Warn().Str("nodeid", nodeidString).Msg("node ID not found in cache. checking database...")
			node, err := srv.core.GetNode(r.Context(), nodeid)

			// authorized is true if there is no error
			authorized = !nodeAuthErr(w, r, err, node)
			if !authorized {
				if err == nil {
					err = errors.New("")
				}
				log.Error().Err(err).Msg("node auth error")
				responses.ErrJsonNodeNotApproved(w, r, nil)
				return
			}
			srv.cache.SetAuth(nodeidString, authorized)
		}

		// at this point, we know fprint validity. Put it in the cache.

		if !authorized {
			return
		}

		if err := srv.core.Seen(r.Context(), nodeid); err != nil {
			responses.ErrJsonServiceUnavailable(w, r, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (srv *SweetToothServer) MiddlewareWebAuth(token string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// tokenString := extractAuthToken(r)
		})
	}
}
