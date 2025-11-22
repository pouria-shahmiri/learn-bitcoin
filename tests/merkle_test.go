package tests

import (
	"testing"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/crypto"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// Test single transaction (base case)
func TestMerkleRootSingleTx(t *testing.T) {
	// Genesis block coinbase transaction hash
	hash, err := types.NewHashFromString(
		"4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
	)
	if err != nil {
		t.Fatal(err)
	}
	
	root := crypto.ComputeMerkleRoot([]types.Hash{hash})
	
	if root != hash {
		t.Error("Single transaction: merkle root should equal transaction hash")
	}
}

// Test even number of transactions
func TestMerkleRootEvenNumber(t *testing.T) {
	// Two real transaction hashes from Bitcoin
	hash1, _ := types.NewHashFromString(
		"4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
	)
	hash2, _ := types.NewHashFromString(
		"0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098",
	)

	root := crypto.ComputeMerkleRoot([]types.Hash{hash1, hash2})
	
	// Manually compute expected root
	combined := append(hash1[:], hash2[:]...)
	expected := crypto.DoubleSHA256(combined)
	
	if root != expected {
		t.Errorf("Got %s, want %s", root, expected)
	}
}

// Test odd number (most important!)
func TestMerkleRootOddNumber(t *testing.T) {
	hash1, _ := types.NewHashFromString(
		"4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
	)
	hash2, _ := types.NewHashFromString(
		"0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098",
	)
	hash3, _ := types.NewHashFromString(
		"9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5",
	)

	root := crypto.ComputeMerkleRoot([]types.Hash{hash1, hash2, hash3})
	
	// Compute expected (hash3 gets duplicated)
	left := crypto.DoubleSHA256(append(hash1[:], hash2[:]...))
	right := crypto.DoubleSHA256(append(hash3[:], hash3[:]...))
	expected := crypto.DoubleSHA256(append(left[:], right[:]...))
	
	if root != expected {
		t.Errorf("Got %s, want %s", root, expected)
	}
	
	t.Logf("Merkle root for 3 transactions: %s", root)
}

// Test tree structure
func TestMerkleTreeStructure(t *testing.T) {
	hashes := make([]types.Hash, 4)
	for i := range hashes {
		hashes[i] = crypto.DoubleSHA256([]byte{byte(i)})
	}

	tree := crypto.BuildMerkleTree(hashes)
	
	// Should have 3 levels for 4 transactions
	// Level 0: 4 hashes
	// Level 1: 2 hashes
	// Level 2: 1 hash (root)
	if len(tree) != 3 {
		t.Errorf("Expected 3 levels, got %d", len(tree))
	}
	
	if len(tree[0]) != 4 {
		t.Error("Level 0 should have 4 hashes")
	}
	if len(tree[1]) != 2 {
		t.Error("Level 1 should have 2 hashes")
	}
	if len(tree[2]) != 1 {
		t.Error("Level 2 should have 1 hash (root)")
	}
}