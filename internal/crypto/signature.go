package crypto

import (
	//"crypto/ed25519"

	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/goodieshq/sweettooth/pkg/info"
	"github.com/google/uuid"
)

const TOKEN_DRIFT_TOLERANCE = 5 * time.Minute
const TOKEN_VALIDITY_PERIOD = 30 * time.Minute
const CLAIM_PUBKEY = "pubkey"

type TokenGenerator func() string

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

func keyFunc(claims jwt.MapClaims, pubkey ed25519.PublicKey, nodeid uuid.UUID) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// extract all necessary claims

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
		if sub != iss || aud != info.APP_NAME {
			return nil, errors.New("unexpected sub/iss/aud values")
		}

		// and that UUID should be the fingerprint derived from pubkey
		if nodeid.String() != sub {
			return nil, errors.New("fingerprint mismatch")
		}

		// we've extracted the public key from the token, now we can verify it
		return pubkey, nil
	}
}

// given a bearer token from an authorization header, verify that it is unexpired, validly formed, and internally consistent
func VerifyNodeJWT(tokenString string) (nodeid uuid.UUID, pubkey ed25519.PublicKey, err error) {
	// parse the token assuming ED25519 key signing
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return
	}

	// extract the claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		err = errors.New("unable to process claims from the token")
		return
	}

	// there should be a pubkey within the claims
	pubkeyB64, err := extractClaim[string](claims, CLAIM_PUBKEY)
	if err != nil {
		return
	}

	// extract the public key bytes from the claim
	pubkeyBytes, err := base64.StdEncoding.DecodeString(pubkeyB64)
	if err != nil {
		return
	}

	// the public key and nodeid should be included in any errors beyond this point.
	pubkey = ed25519.PublicKey(pubkeyBytes)
	nodeid = Fingerprint(pubkey)

	// re-parse the token now that we know the expected public key
	token, err = jwt.ParseWithClaims(tokenString, claims, keyFunc(claims, pubkey, nodeid))

	// ensure token was parsed and is valid
	if err != nil || !token.Valid {
		if err == nil {
			err = errors.New("unable to parse token")
		}
		return
	}
	err = nil
	return
}
