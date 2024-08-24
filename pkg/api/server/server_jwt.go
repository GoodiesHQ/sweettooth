package server

import (
	"net/http"
	"strings"

	"github.com/goodieshq/sweettooth/internal/crypto"
	"github.com/goodieshq/sweettooth/internal/util"
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

// Middleware for handling endpoints which are exclusively used by nodes running sweettooth clients
func (srv *SweetToothServer) MiddlewareNodeAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// extract the bearer token from the Authorization header
		tokenString := extractAuthToken(r)
		if tokenString == "" {
			ErrNodeTokenInvalid(w, r, nil)
			return
		}

		nodeid, _, err := crypto.VerifyNodeJWT(tokenString)
		nodeidString := nodeid.String()
		if nodeidString != "" {
			// set the node ID (may be returned even upon error)
			util.SetRequestNodeID(r, nodeid)
		}
		if err != nil {
			ErrNodeTokenInvalid(w, r, err)
		}

		/* At this point, all we know is that the signature is valid and well-formed. Check cache/db for validity */
		_, found := srv.cacheValidNodeIDs.Get(nodeidString)
		if !found {
			node, err := srv.core.GetNode(r.Context(), nodeid)
			if err != nil {
				ErrServiceUnavailable(w, r, err)
				return
			}

			if node == nil {
				ErrNodeNotFound(w, r, nil)
				return
			}

			if !node.Approved {
				ErrNodeNotApproved(w, r, nil)
				return
			}
		}

		// at this point, we know fprint is valid. Put it in the cache
		srv.cacheValidNodeIDs.Set(nodeidString, true, 0)

		if err := srv.core.Seen(r.Context(), nodeid); err != nil {
			ErrServiceUnavailable(w, r, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}
