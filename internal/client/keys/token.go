package keys

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/goodieshq/sweettooth/internal/crypto"
	"github.com/goodieshq/sweettooth/pkg/info"
)

// Create a JWT signed by the node's own key
func CreateNodeJWT() (string, error) {
	now := time.Now().UTC()
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, jwt.MapClaims{
		"iss":               GetPublicKeyID().String(),                                                      // issuer is the current node
		"sub":               GetPublicKeyID().String(),                                                      // subject is the current node
		"aud":               info.APP_NAME,                                                                  // audience should be the app server
		"iat":               now.Unix(),                                                                     // issued just now
		"nbf":               now.Add(-crypto.TOKEN_DRIFT_TOLERANCE).Unix(),                                  // allow some clock drift tolerance, up to you how much
		"exp":               now.Add(crypto.TOKEN_VALIDITY_PERIOD).Add(crypto.TOKEN_DRIFT_TOLERANCE).Unix(), // add expiration time plus the clock drift tolerance
		crypto.CLAIM_PUBKEY: GetPublicKeyBase64(),                                                           // public key should be included in each request and checked against the iss/sub
	})
	return token.SignedString(getSecretKey())
}
