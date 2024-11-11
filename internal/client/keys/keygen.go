package keys

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"sync"

	"github.com/goodieshq/sweettooth/internal/crypto"
	"github.com/goodieshq/sweettooth/internal/util"
	"github.com/goodieshq/sweettooth/pkg/config"
	"github.com/rs/zerolog/log"
)

var keyMu sync.Mutex

// Checkes whether the signing keys have already been generated. These will be used for authentication tokens
func keysExist() bool {
	// ensure both public and private keys exist
	return util.IsFile(config.SecretKey()) && util.IsFile(config.PublicKey())
}

// load the keys from disk into memory
func loadKeys(cipher crypto.Cipher) error {
	keyMu.Lock()
	defer keyMu.Unlock()

	var err error

	// decrypt the contents of the file using the windows DBAPI
	secPem, err := crypto.DecipherFile(cipher, config.SecretKey())
	if err != nil {
		return err
	}

	// decode the pem block and parse the private key
	secBlock, _ := pem.Decode(secPem)
	secKeyBytes, err := x509.ParsePKCS8PrivateKey(secBlock.Bytes)
	if err != nil {
		return err
	}

	// convert the extracted bytes to an ED25519 private key
	secKey, ok := secKeyBytes.(ed25519.PrivateKey)
	if !ok {
		return fmt.Errorf("unable to convert to ed25519.PrivateKey")
	}

	// decode the contents of the file, do not use cipher for the public key
	pubPem, err := os.ReadFile(config.PublicKey())
	if err != nil {
		return err
	}

	// decode the pem block and parse the public key
	pubBlock, _ := pem.Decode(pubPem)
	pubKeyBytes, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return err
	}

	// convert the extracted bytes to an ED25519 public key
	pubKey, ok := pubKeyBytes.(ed25519.PublicKey)
	if !ok {
		return fmt.Errorf("unable to convert to ed25519.PublicKey")
	}

	// set the application-wide keypair
	return SetKeyPair(pubKey, secKey)
}

// securely export the keys to disk
func saveKeys(cipher crypto.Cipher, pubKey ed25519.PublicKey, secKey ed25519.PrivateKey) error {
	keyMu.Lock()
	defer keyMu.Unlock()

	if err := SetKeyPair(pubKey, secKey); err != nil {
		return err
	}

	/*
	 * secret key - encode, encrypt, export
	 */
	secBytes, err := x509.MarshalPKCS8PrivateKey(secKey)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal secret key")
		return err
	}

	secBlock := &pem.Block{Type: "PRIVATE KEY", Bytes: secBytes}
	secEncoded := pem.EncodeToMemory(secBlock)

	err = crypto.EncipherFile(cipher, config.SecretKey(), secEncoded)
	if err != nil {
		log.Error().Err(err).Msg("failed to encrypt the key")
		return err
	}

	/*
	 * public key - encode, export
	 */
	pubBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal public key")
		return err
	}

	pubBlock := &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}
	pubEncoded := pem.EncodeToMemory(pubBlock)

	err = os.WriteFile(config.PublicKey(), pubEncoded, 0644)
	if err != nil {
		log.Error().Err(err).Msg("failed to write the client public key")
		return err
	}

	return nil
}

func generate(cipher crypto.Cipher) error {
	// generate a new asymmetric keypair
	pubKey, secKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	return saveKeys(cipher, pubKey, secKey)
}

func Bootstrap(cipher crypto.Cipher) error {
	// create new or load existing keys (ideally only done once per system)
	if !keysExist() {
		return generate(cipher)
	}
	return loadKeys(cipher)
}
