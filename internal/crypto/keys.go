package crypto

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
)

// store the keys used throughout the application's lifespan
var nodePublicKey ed25519.PublicKey
var nodeSecretKey ed25519.PrivateKey
var nodePublicKeyID uuid.UUID

func SetKeyPair(publicKey ed25519.PublicKey, secretKey ed25519.PrivateKey) error {
	// perform a test to ensure the keys are valid
	sig := ed25519.Sign(secretKey, publicKey)
	if ed25519.Verify(publicKey, publicKey, sig) {
		// keys are valid, store application-wide keys
		nodePublicKey = publicKey
		nodeSecretKey = secretKey
		nodePublicKeyID = Fingerprint(publicKey)
		return nil
	}
	return fmt.Errorf("public and Private key mismatch")
}

func GetPublicKey() ed25519.PublicKey {
	return nodePublicKey
}

// UUID fingerprint
func GetPublicKeyID() uuid.UUID {
	return nodePublicKeyID
}

// sign the node's public key itself using the node's private key
func GetPublicKeySig() []byte {
	return Sign(GetPublicKey())
}

// sign the node's public key itself using the node's private key, convert to base64
func GetPublicKeySigBase64() string {
	return base64.StdEncoding.EncodeToString(GetPublicKeySig())
}

// acquire the node's public key in base64 format
func GetPublicKeyBase64() string {
	return base64.StdEncoding.EncodeToString(nodePublicKey)
}

// create a signature from a chunk of data signed by the node's private key
func Sign(data []byte) []byte {
	return ed25519.Sign(getSecretKey(), data)
}

// verify a block of data was signed with a public key given a raw signature
func Verify(pubkey ed25519.PublicKey, data, sig []byte) bool {
	return ed25519.Verify(pubkey, data, sig)
}

// verify a block of data was signed with a public key given a base64 signature
func VerifyBase64(pubkey ed25519.PublicKey, data []byte, sigBase64 string) bool {
	sig, err := base64.StdEncoding.DecodeString(sigBase64)
	return err == nil && Verify(pubkey, data[:], sig)
}

// aquire the node's stored secret key
func getSecretKey() ed25519.PrivateKey {
	return nodeSecretKey
}
