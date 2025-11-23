package keys

import (
	"crypto/sha256"
	"fmt"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"golang.org/x/crypto/ripemd160"
)

// PublicKey represents a Bitcoin public key
type PublicKey struct {
	key *secp256k1.PublicKey
}

// Bytes returns serialized public key
func (pub *PublicKey) Bytes(compressed bool) []byte {
	if compressed {
		return pub.key.SerializeCompressed()
	}
	return pub.key.SerializeUncompressed()
}

// Hash160 returns RIPEMD160(SHA256(pubkey))
// This is used for address generation
func (pub *PublicKey) Hash160() []byte {
	// Step 1: SHA-256
	sha := sha256.Sum256(pub.Bytes(true))

	// Step 2: RIPEMD-160
	ripe := ripemd160.New()
	ripe.Write(sha[:])

	return ripe.Sum(nil)
}

// String returns hex representation
func (pub *PublicKey) String() string {
	return fmt.Sprintf("%x", pub.Bytes(true))
}

// IsCompressed checks if public key is in compressed format
func (pub *PublicKey) IsCompressed() bool {
	// Compressed keys start with 0x02 or 0x03
	serialized := pub.key.SerializeCompressed()
	return serialized[0] == 0x02 || serialized[0] == 0x03
}

// Verify verifies a signature against a message hash
func (pub *PublicKey) Verify(hash []byte, sig *Signature) bool {
	if len(hash) != 32 {
		return false
	}

	return sig.sig.Verify(hash, pub.key)
}
