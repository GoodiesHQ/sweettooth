package crypto

import (
	"crypto/ed25519"

	"github.com/google/uuid"
)

// Official SweetTooth namespace
var namspaceSweetToothPublicKeys = uuid.MustParse("7d2923a0-877c-4eb2-9df6-d89a807cd923")

// Generate a fingerprint from a public key (used as UUID primary key in database as endpoint identifier)
func Fingerprint(publicKey ed25519.PublicKey) uuid.UUID {
	return uuid.NewSHA1(namspaceSweetToothPublicKeys, publicKey)
}
