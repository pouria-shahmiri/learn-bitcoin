package tests

import (
	"testing"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/serialization"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// Test genesis block header
func TestGenesisBlockHeader(t *testing.T) {
	// Bitcoin's genesis block
	header := &types.BlockHeader{
		Version:       1,
		PrevBlockHash: types.Hash{}, // All zeros (first block)
		MerkleRoot: mustHash(
			"4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
		).Reverse(),
		Timestamp: 1231006505, // Jan 3, 2009
		Bits:      0x1d00ffff, // Initial difficulty
		Nonce:     2083236893, // Winning nonce
	}

	hash, err := serialization.HashBlockHeader(header)
	if err != nil {
		t.Fatal(err)
	}

	// Genesis block hash (note: displayed in reverse byte order)
	// Internal: 6fe28c0ab6f1b372c1a6a246ae63f74f931e8365e15a089c68d6190000000000
	// Displayed: 000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f
	
	expectedHash := "6fe28c0ab6f1b372c1a6a246ae63f74f931e8365e15a089c68d6190000000000"
	
	t.Logf("Genesis block hash: %s", hash)
	t.Logf("Expected: %s", expectedHash)
	
	// Verify it starts with many zeros (low difficulty in 2009)
	if !hasLeadingZeros(hash, 5) {
		t.Error("Genesis hash should have leading zeros")
	}
}

// Test block header serialization is exactly 80 bytes
func TestBlockHeaderSize(t *testing.T) {
	header := &types.BlockHeader{
		Version:       1,
		PrevBlockHash: types.Hash{},
		MerkleRoot:    types.Hash{},
		Timestamp:     1231006505,
		Bits:          0x1d00ffff,
		Nonce:         0,
	}

	serialized, err := serialization.SerializeBlockHeader(header)
	if err != nil {
		t.Fatal(err)
	}

	if len(serialized) != 80 {
		t.Errorf("Block header must be 80 bytes, got %d", len(serialized))
	}
}

// Test changing nonce changes hash
func TestNonceChangesHash(t *testing.T) {
	header := &types.BlockHeader{
		Version:       1,
		PrevBlockHash: types.Hash{},
		MerkleRoot:    types.Hash{},
		Timestamp:     1231006505,
		Bits:          0x1d00ffff,
		Nonce:         0,
	}

	hash1, _ := serialization.HashBlockHeader(header)
	
	header.Nonce = 1
	hash2, _ := serialization.HashBlockHeader(header)

	if hash1 == hash2 {
		t.Error("Changing nonce should change hash")
	}
}

// Helper functions
func mustHash(s string) types.Hash {
	h, err := types.NewHashFromString(s)
	if err != nil {
		panic(err)
	}
	return h
}

func hasLeadingZeros(h types.Hash, count int) bool {
	for i := len(h) - 1; i >= len(h)-count; i-- {
		if h[i] != 0 {
			return false
		}
	}
	return true
}