package crypto

import (
	"crypto/ed25519"
	"encoding/base64"
)

// Use ed25519 signature algorithm to sign the payload
func Sign(message []byte) []byte {
	return ed25519.Sign(GetSecretKey(), message)
}

func SignBase64(message []byte) string {
	return base64.StdEncoding.EncodeToString(message)
}

// Verification that the message is signed by the client
func VerifyClient(message []byte, signature []byte) bool {
	return ed25519.Verify(GetPublicKey(), message, signature)
}

func VerifyClientBase64(message []byte, signature string) (bool, error) {
	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, err
	}
	return VerifyClient(message, sig), nil
}

// Verification that the message is signed by the server
func VerifyServer(message []byte, signature []byte) bool {
	// TODO: get server public key instead of our own
	return ed25519.Verify(GetPublicKey(), message, signature)
}

func VerifyServerBase64(message []byte, signature string) (bool, error) {
	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, err
	}
	return VerifyServer(message, sig), nil
}
