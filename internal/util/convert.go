package util

import (
	"crypto/ed25519"
	"encoding/base64"
)

func Base64toBytes(bytesBase64 string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(bytesBase64)
}

func Base64toPubKey(pubkeyBase64 string) (ed25519.PublicKey, error) {
	bytes, err := Base64toBytes(pubkeyBase64)
	if err != nil {
		return nil, err
	}
	return ed25519.PublicKey(bytes), nil
}
