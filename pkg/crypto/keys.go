package crypto

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"

	"github.com/rs/zerolog/log"
)

// store the keys used throughout the application's lifespan
var nodePublicKey ed25519.PublicKey
var nodeSecretKey ed25519.PrivateKey

func SetKeyPair(publicKey ed25519.PublicKey, secretKey ed25519.PrivateKey) error {
	// perform a test to ensure the keys are valid
	var signme = []byte("sign me")

	sig := ed25519.Sign(secretKey, signme)
	if ed25519.Verify(publicKey, signme, sig) {
		// keys are valid, store application-wide keys
		nodePublicKey = publicKey
		nodeSecretKey = secretKey
		log.Debug().Msg("Loaded keys in memory")
		return nil
	}
	return fmt.Errorf("public and Private key mismatch")
}

func GetPublicKey() ed25519.PublicKey {
	return nodePublicKey
}

func GetPublicKeyBase64() string {
	return base64.StdEncoding.EncodeToString(nodePublicKey)
}

func GetSecretKey() ed25519.PrivateKey {
	return nodeSecretKey
}
