// This is the fundamental building block. Bitcoin uses 32-byte hashes everywhere. :DDD

package types

import (
	"encoding/hex"
	"fmt"
)

// Hash represents a 32-byte hash (SHA-256 output)
type Hash [32]byte

// String returns hex representation for printing
func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

// NewHashFromString creates hash from hex string
// Used for testing with known Bitcoin hashes
func NewHashFromString(s string) (Hash, error) {
	var h Hash
	b, err := hex.DecodeString(s)
	if err != nil {
		return h, fmt.Errorf("invalid hex: %w", err)
	}
	if len(b) != 32 {
		return h, fmt.Errorf("hash must be 32 bytes, got %d", len(b))
	}
	copy(h[:], b)
	return h, nil
}

// IsZero checks if hash is all zeros (used for genesis block)
func (h Hash) IsZero() bool {
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return true
}

// Reverse returns a new Hash with bytes reversed
func (h Hash) Reverse() Hash {
	var reversed Hash
	for i := 0; i < 32; i++ {
		reversed[i] = h[31-i]
	}
	return reversed
}