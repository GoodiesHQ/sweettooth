package server

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/goodieshq/sweettooth/pkg/config"
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

// Middleware for handling endpoints which are exclusively used by nodes to interact with
func (srv *SweetToothServer) MiddlewareNodeAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// extract the bearer token from the Authorization header
		tokenString := extractAuthToken(r)
		if tokenString == "" {
			ErrNodeTokenInvalid(w, r, nil)
			return
		}

		// parse the token without verification to extract the public key
		token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
		if err != nil {
			ErrNodeTokenInvalid(w, r, err)
			return
		}

		// extract the claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			ErrNodeTokenInvalid(w, r, err)
			return
		}

		// there should be a pubkey within the claims
		pubkeyB64, err := extractClaim[string](claims, "pubkey")
		if err != nil {
			ErrNodeTokenInvalid(w, r, err)
			return
		}

		pubkeyBytes, err := base64.StdEncoding.DecodeString(pubkeyB64)
		if err != nil {
			ErrNodeTokenInvalid(w, r, err)
			return
		}

		pubkey := ed25519.PublicKey(pubkeyBytes)
		fprint := crypto.Fingerprint(pubkey)
		fprints := fprint.String()
		util.SetRequestNodeID(r, fprints)

		token, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// extract the necessary claims
			sub, err := extractClaim[string](claims, "sub")
			if err != nil {
				return nil, err
			}

			iss, err := extractClaim[string](claims, "iss")
			if err != nil {
				return nil, err
			}

			aud, err := extractClaim[string](claims, "aud")
			if err != nil {
				return nil, err
			}

			// for node tokens, the iss and sub should both be the UUID
			if sub != iss || aud != config.APP_NAME {
				return nil, errors.New("unexpected sub/iss/aud values")
			}

			// and that UUID should be the fingerprint derived from pubkey
			if fprints != sub {
				return nil, errors.New("fingerprint mismatch")
			}

			return pubkey, nil
		})

		if err != nil || !token.Valid {
			ErrNodeTokenInvalid(w, r, err)
			return
		}

		/* At this point, all we know is that the signature is valid and well-formed. Check cache/db for validity */
		_, found := srv.cacheValidNodeIDs.Get(fprints)
		if !found {
			node, err := srv.core.GetNodeByID(fprint)
			if err != nil {
				ErrNodeTokenInvalid(w, r, err)
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

		srv.cacheValidNodeIDs.Set(fprints, true, 0)

		next.ServeHTTP(w, r)
	})
}
