package crypto

import (
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// ComputeMerkleRoot calculates root from transaction hashes
func ComputeMerkleRoot(txHashes []types.Hash) types.Hash {
	// Edge case: empty block (shouldn't happen)
	if len(txHashes) == 0 {
		return types.Hash{}
	}

	// Start with transaction hashes as base level
	currentLevel := make([]types.Hash, len(txHashes))
	copy(currentLevel, txHashes)

	// Keep combining pairs until one hash remains
	for len(currentLevel) > 1 {
		var nextLevel []types.Hash

		// Process pairs
		for i := 0; i < len(currentLevel); i += 2 {
			left := currentLevel[i]

			var right types.Hash
			if i+1 < len(currentLevel) {
				// Normal case: pair exists
				right = currentLevel[i+1]
			} else {
				// Odd number: duplicate last hash
				right = currentLevel[i]
			}

			// Concatenate left + right, then hash
			combined := append(left[:], right[:]...)
			parentHash := DoubleSHA256(combined)
			nextLevel = append(nextLevel, parentHash)
		}

		currentLevel = nextLevel
	}

	return currentLevel[0]
}

// BuildMerkleTree returns all levels (for debugging/visualization)
func BuildMerkleTree(txHashes []types.Hash) [][]types.Hash {
	if len(txHashes) == 0 {
		return nil
	}

	var tree [][]types.Hash
	currentLevel := make([]types.Hash, len(txHashes))
	copy(currentLevel, txHashes)

	// Add base level
	tree = append(tree, currentLevel)

	// Build upward
	for len(currentLevel) > 1 {
		var nextLevel []types.Hash

		for i := 0; i < len(currentLevel); i += 2 {
			left := currentLevel[i]
			right := currentLevel[i]
			if i+1 < len(currentLevel) {
				right = currentLevel[i+1]
			}

			combined := append(left[:], right[:]...)
			nextLevel = append(nextLevel, DoubleSHA256(combined))
		}

		tree = append(tree, nextLevel)
		currentLevel = nextLevel
	}

	return tree
}

/*
```
**Visual explanation:**
```
4 transactions: [A, B, C, D]

Level 0: [H(A), H(B), H(C), H(D)]  ← Transaction hashes
         /  \    /  \
Level 1: [H(AB), H(CD)]             ← Pair them and hash
           \     /
Level 2:   [H(ABCD)]                ← Merkle Root
```

**With odd number (3 transactions):**
```
Level 0: [H(A), H(B), H(C)]
         /  \    /  \
Level 1: [H(AB), H(CC)]  ← C duplicated!
           \     /
Level 2:   [H(ABCC)]
```

**Why Merkle trees?**

1. **Efficiency:** Prove a transaction is in a block with only log(n) hashes
2. **SPV (Simplified Payment Verification):** Mobile wallets don't need full blockchain
3. **Bandwidth:** Light clients download headers only (80 bytes each)

**Example proof:**
```
To prove transaction B is in block:
- Download: H(A), H(CD), merkle root
- Compute: H(AB) = H(H(A) + H(B))
- Compute: H(ABCD) = H(H(AB) + H(CD))
- Compare: Does H(ABCD) match merkle root?

*/
