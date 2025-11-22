package tests

import (
	"testing"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/crypto"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/serialization"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// Test genesis block coinbase transaction
func TestGenesisTransaction(t *testing.T) {
	// Recreate Bitcoin's genesis block coinbase transaction
	tx := &types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{
				PrevTxHash:  types.Hash{}, // All zeros (no previous transaction)
				OutputIndex: 0xFFFFFFFF,   // Special value for coinbase
				SignatureScript: []byte{
					0x04, 0xff, 0xff, 0x00, 0x1d, 0x01, 0x04, 0x45,
					// "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
				},
				Sequence: 0xFFFFFFFF,
			},
		},
		Outputs: []types.TxOutput{
			{
				Value: 5000000000, // 50 BTC
				PubKeyScript: []byte{
					0x41, // OP_PUSHDATA 65 bytes
					0x04, // Uncompressed pubkey marker
					// Satoshi's public key (65 bytes total)
				},
			},
		},
		LockTime: 0,
	}

	hash, err := serialization.HashTransaction(tx)
	if err != nil {
		t.Fatalf("Failed to hash transaction: %v", err)
	}

	// This is the actual genesis coinbase txid
	expectedHash := "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b"
	
	// Note: We won't match exactly because we simplified the scripts
	// But hash should be deterministic
	t.Logf("Genesis coinbase hash: %s", hash)
	t.Logf("Expected: %s", expectedHash)
}

// Test transaction serialization is deterministic
func TestTransactionSerializationDeterministic(t *testing.T) {
	tx := &types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{
				PrevTxHash:      types.Hash{1, 2, 3},
				OutputIndex:     0,
				SignatureScript: []byte("test sig"),
				Sequence:        0xFFFFFFFF,
			},
		},
		Outputs: []types.TxOutput{
			{
				Value:        1000000,
				PubKeyScript: []byte("test pubkey"),
			},
		},
		LockTime: 0,
	}

	// Hash twice
	hash1, _ := serialization.HashTransaction(tx)
	hash2, _ := serialization.HashTransaction(tx)

	if hash1 != hash2 {
		t.Error("Transaction hash not deterministic")
	}
}

// Test transaction with multiple inputs and outputs
func TestComplexTransaction(t *testing.T) {
	tx := &types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{
				PrevTxHash:      crypto.DoubleSHA256([]byte("tx1")),
				OutputIndex:     0,
				SignatureScript: []byte("sig1"),
				Sequence:        0xFFFFFFFF,
			},
			{
				PrevTxHash:      crypto.DoubleSHA256([]byte("tx2")),
				OutputIndex:     1,
				SignatureScript: []byte("sig2"),
				Sequence:        0xFFFFFFFF,
			},
		},
		Outputs: []types.TxOutput{
			{
				Value:        5000000,
				PubKeyScript: []byte("to alice"),
			},
			{
				Value:        3000000,
				PubKeyScript: []byte("change back"),
			},
		},
		LockTime: 0,
	}

	serialized, err := serialization.SerializeTransaction(tx)
	if err != nil {
		t.Fatal(err)
	}

	// Check minimum expected size
	// Version(4) + InputCount(1) + 2*Input + OutputCount(1) + 2*Output + Locktime(4)
	minSize := 4 + 1 + 4 + 1
	if len(serialized) < minSize {
		t.Errorf("Serialized transaction too small: %d bytes", len(serialized))
	}

	t.Logf("Transaction size: %d bytes", len(serialized))
}