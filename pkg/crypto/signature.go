package crypto

import (
	//"crypto/ed25519"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/goodieshq/sweettooth/pkg/config"
)

const TOKEN_DRIFT_TOLERANCE = 5 * time.Minute
const TOKEN_VALIDITY_PERIOD = 30 * time.Minute

type TokenGenerator func() string

// Create a JWT signed by the node's own key
func CreateNodeJWT() (string, error) {
	now := time.Now().UTC()
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, jwt.MapClaims{
		// issuer and subject are both the current node, audience is the sweettooth application server
		"iss":    GetPublicKeyID(),
		"sub":    GetPublicKeyID(),
		"pubkey": GetPublicKeyBase64(),
		"iat":    now.Unix(),
		"nbf":    now.Add(-TOKEN_DRIFT_TOLERANCE).Unix(),
		"exp":    now.Add(TOKEN_VALIDITY_PERIOD).Add(TOKEN_DRIFT_TOLERANCE).Unix(),
		"aud":    config.APP_NAME,
	})
	return token.SignedString(GetSecretKey())
}

func VerifyNodeJWT(token string) error {
	tok, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return 1, nil
	})
	if err != nil {
		return err
	}

	if tok == nil {
		return nil
	}
	return nil
}
