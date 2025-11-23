package crypto

import (
	"crypto/sha256"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// DoubleSHA256 is Bitcoin's hash function
// Why double? Prevents length extension attacks
func DoubleSHA256(data []byte) types.Hash {
	// First hash
	firstHash := sha256.Sum256(data)

	// Hash the hash
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash
}

// HashTransaction computes transaction ID (txid)
func HashTransaction(data []byte) types.Hash {
	return DoubleSHA256(data)
}

// HashBlockHeader computes block hash
func HashBlockHeader(data []byte) types.Hash {
	return DoubleSHA256(data)
}

/*

**Why double hashing?**

Single SHA-256 has a vulnerability called "length extension attack":
- Attacker can append data to a message and compute valid hash
- Double hashing prevents this
- Bitcoin developers chose this for extra security

**How it works:**
Input data → SHA-256 → intermediate hash → SHA-256 → final hash
                       (32 bytes)                      (32 bytes)

*/
