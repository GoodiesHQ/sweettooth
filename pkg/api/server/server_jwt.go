package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/goodieshq/sweettooth/pkg/crypto"
	"github.com/goodieshq/sweettooth/pkg/util"
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

// extract the target claim from the token's claims and cast to type T
func extractClaim[T any](claims jwt.MapClaims, name string) (T, error) {
	var zero T

	// check if the claim exists first
	key, found := claims[name]
	if !found {
		return zero, fmt.Errorf("claim '%s' not found", name)
	}

	// convert the value to type T if possible
	val, ok := key.(T)
	if !ok {
		return zero, fmt.Errorf("unexpected type for claim '%s'", name)
	}

	return val, nil
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
