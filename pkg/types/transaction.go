package types

import (
	"fmt"
)

// TxInput represents where coins come from
type TxInput struct {
	PrevTxHash      Hash   // Which transaction created these coins?
	OutputIndex     uint32 // Which output in that transaction?
	SignatureScript []byte // Proof you can spend (signature + pubkey)
	Sequence        uint32 // For timelock features (usually 0xFFFFFFFF)
}

// TxOutput represents where coins go
type TxOutput struct {
	Value        int64  // Amount in satoshis (1 BTC = 100,000,000 satoshis)
	PubKeyScript []byte // Conditions to spend (usually "pay to this address")
}

// OutPoint uniquely identifies a transaction output
type OutPoint struct {
	Hash  Hash   // Transaction hash
	Index uint32 // Output index
}

// NewOutPoint creates a new outpoint
func NewOutPoint(hash Hash, index uint32) OutPoint {
	return OutPoint{
		Hash:  hash,
		Index: index,
	}
}

// String returns string representation of outpoint
func (op OutPoint) String() string {
	return fmt.Sprintf("%s:%d", op.Hash, op.Index)
}

// Equal checks if two outpoints are equal
func (op OutPoint) Equal(other OutPoint) bool {
	return op.Hash == other.Hash && op.Index == other.Index
}

// Transaction is a value transfer
type Transaction struct {
	Version  int32      // Protocol version
	Inputs   []TxInput  // Where coins come from
	Outputs  []TxOutput // Where coins go
	LockTime uint32     // When tx becomes valid (0 = immediately)
}

/*
**Key concepts explained:**

1. **TxInput:**
   - `PrevTxHash`: Points to the transaction that created the coins you're spending
   - `OutputIndex`: Which output from that transaction (transactions can have multiple outputs)
   - `SignatureScript`: Your proof that you own those coins (we'll add signatures in Milestone 3)
   - `Sequence`: Advanced feature for replacing transactions

2. **TxOutput:**
   - `Value`: Amount in satoshis (smallest Bitcoin unit)
   - `PubKeyScript`: A small program that defines how coins can be spent (like "must have signature from this address")

3. **Transaction:**
   - Links inputs (what you're spending) to outputs (where it goes)
   - The difference between input and output values is the miner fee

**Example in your mind:**


Alice has 10 BTC from transaction ABC (output #0)
She wants to send 7 BTC to Bob

Transaction:
  Input: Previous TX = ABC, Output Index = 0
  Output 1: 7 BTC to Bob
  Output 2: 3 BTC back to Alice (change)

*/
