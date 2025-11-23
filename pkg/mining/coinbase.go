package mining

import (
	"encoding/binary"
	"fmt"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/script"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/validation"
)

// CreateCoinbase creates a coinbase transaction for a block
func CreateCoinbase(blockHeight uint64, totalFees int64, minerAddress string, extraNonce uint64) (*types.Transaction, error) {
	// Calculate block reward
	blockReward := validation.GetBlockReward(blockHeight)
	totalReward := blockReward + totalFees

	// Create coinbase input
	// Coinbase input has:
	// - PrevTxHash: all zeros
	// - OutputIndex: 0xFFFFFFFF
	// - SignatureScript: arbitrary data (includes block height per BIP34)
	coinbaseInput := types.TxInput{
		PrevTxHash:      types.Hash{}, // All zeros
		OutputIndex:     0xFFFFFFFF,
		SignatureScript: createCoinbaseScript(blockHeight, extraNonce),
		Sequence:        0xFFFFFFFF,
	}

	// Create output to miner
	// For simplicity, we'll use a P2PKH script
	// In a real implementation, we'd decode the address to get the pubkey hash
	// For now, use a dummy 20-byte hash derived from the address
	pubKeyHash := make([]byte, 20)
	copy(pubKeyHash, []byte(minerAddress))

	minerScript, err := script.P2PKH(pubKeyHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create miner script: %w", err)
	}

	coinbaseOutput := types.TxOutput{
		Value:        totalReward,
		PubKeyScript: minerScript,
	}

	// Create transaction
	coinbaseTx := &types.Transaction{
		Version:  1,
		Inputs:   []types.TxInput{coinbaseInput},
		Outputs:  []types.TxOutput{coinbaseOutput},
		LockTime: 0,
	}

	return coinbaseTx, nil
}

// createCoinbaseScript creates the coinbase scriptSig
// BIP34 requires block height to be first item
func createCoinbaseScript(blockHeight uint64, extraNonce uint64) []byte {
	// Encode block height (BIP34)
	heightBytes := encodeBlockHeight(blockHeight)

	// Add extra nonce for additional entropy
	nonceBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(nonceBytes, extraNonce)

	// Combine: [height_length][height][nonce]
	script := make([]byte, 0, 1+len(heightBytes)+len(nonceBytes))
	script = append(script, byte(len(heightBytes))) // Push height length
	script = append(script, heightBytes...)
	script = append(script, nonceBytes...)

	// Add some arbitrary data (miner can put anything here)
	minerTag := []byte("learn-bitcoin-miner")
	script = append(script, minerTag...)

	return script
}

// encodeBlockHeight encodes block height in minimal format
func encodeBlockHeight(height uint64) []byte {
	if height == 0 {
		return []byte{0}
	}

	// Encode as little-endian, minimal bytes
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], height)

	// Find minimal length
	length := 8
	for length > 0 && buf[length-1] == 0 {
		length--
	}

	return buf[:length]
}

/*
**Coinbase Transaction Explained:**

1. **What is Coinbase?**
   - First transaction in every block
   - Creates new bitcoins (block reward)
   - Collects transaction fees
   - No inputs from previous transactions

2. **Coinbase Input:**
   - PrevTxHash: All zeros (no previous transaction)
   - OutputIndex: 0xFFFFFFFF (special marker)
   - SignatureScript: Arbitrary data (miner can put anything)
   - Must include block height (BIP34) to prevent duplicate coinbase txs

3. **Block Reward:**
   - Started at 50 BTC (2009)
   - Halves every 210,000 blocks (~4 years)
   - Current: 6.25 BTC (as of 2024)
   - Eventually: 0 BTC (miners only get fees)

4. **BIP34:**
   - Requires block height in coinbase scriptSig
   - Prevents duplicate coinbase transactions
   - Activated in 2012

**Example:**
```
Block 100:
  Reward: 50 BTC
  Fees: 0.5 BTC
  Total: 50.5 BTC to miner

Block 210,000:
  Reward: 25 BTC (first halving!)
  Fees: 1 BTC
  Total: 26 BTC to miner
```

**Visual:**
```
Coinbase Transaction:
  Input:
    PrevTx: 0000...0000
    Index: 0xFFFFFFFF
    Script: [height][extra_nonce][miner_tag]

  Output:
    Value: 50.5 BTC
    Script: Pay to miner's address
```
*/
