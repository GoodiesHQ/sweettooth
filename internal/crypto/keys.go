package crypto

import (
	"crypto/ed25519"
	"encoding/base64"
)

// verify a block of data was signed with a public key given a raw signature
func Verify(pubkey ed25519.PublicKey, data, sig []byte) bool {
	return ed25519.Verify(pubkey, data, sig)
}

// verify a block of data was signed with a public key given a base64 signature
func VerifyBase64(pubkey ed25519.PublicKey, data []byte, sigBase64 string) bool {
	sig, err := base64.StdEncoding.DecodeString(sigBase64)
	return err == nil && Verify(pubkey, data[:], sig)
}
