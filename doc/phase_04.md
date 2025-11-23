# Phase 4: Transactions

## 1. Objectives
Phase 4 is a major leap. We move away from storing arbitrary strings ("Send 1 BTC") to a real **Transaction** system. Bitcoin uses the **UTXO (Unspent Transaction Output)** model, not a simple "Account Balance" model like a bank.

## 2. Key Concepts

### UTXO Model vs. Account Model
*   **Account Model (Bank)**: Database says "Alice: $50". Spending $10 updates it to "Alice: $40".
*   **UTXO Model (Bitcoin)**:
    *   Alice receives 50 BTC. This is an **Output** locked to her.
    *   To spend 10 BTC, she creates an **Input** that consumes the entire 50 BTC output.
    *   She creates two new **Outputs**: 10 BTC to Bob, and 40 BTC back to herself (Change).
    *   **Balance**: The sum of all unspent outputs you can unlock.

### Transaction Structure
A transaction consists of:
1.  **ID**: Hash of the transaction.
2.  **Inputs (Vin)**: References to previous outputs (`TxID`, `Vout`) and a signature (`ScriptSig`) to prove ownership.
3.  **Outputs (Vout)**: New coins created (`Value`) and a puzzle (`ScriptPubKey`) that the receiver must solve to spend them.

### Coinbase Transaction
The first transaction in a block. It has no inputs (creates coins from thin air) and rewards the miner.

## 3. Code Implementation

### `Transaction` Struct
Located in `pkg/transaction/transaction.go`.

```go
type Transaction struct {
    ID   []byte
    Vin  []TXInput
    Vout []TXOutput
}
```

### `TXInput`
References a previous output.

```go
type TXInput struct {
    Txid      []byte // ID of the transaction we are spending
    Vout      int    // Index of the output in that transaction
    Signature []byte // Proof we own it (simplified in Phase 4)
    PubKey    []byte // Our public key
}
```

### `TXOutput`
Locks coins to a puzzle (usually a public key hash).

```go
type TXOutput struct {
    Value        int    // Amount in Satoshis
    PubKeyHash   []byte // Hash of the receiver's public key
}
```

## 4. Architecture Diagram

### Transaction Chain
How money moves from Alice to Bob.

```text
Transaction 1 (Coinbase)
+-----------------------------+
| ID: 1111...                 |
| In: None                    |
| Out 0: 50 BTC -> Alice      |
+-----------------------------+
            |
            | (Alice spends Out 0)
            v
Transaction 2 (Spending)
+-----------------------------+
| ID: 2222...                 |
| In: Ref Tx1, Out 0          |
| Out 0: 10 BTC -> Bob        |
| Out 1: 40 BTC -> Alice (Chg)|
+-----------------------------+
```

## 5. How to Run
The CLI commands change to support addresses and amounts.

```bash
# Create a wallet/blockchain
go run cmd/phase_4/main.go createblockchain -address "Alice"

# Send coins
go run cmd/phase_4/main.go send -from "Alice" -to "Bob" -amount 10

# Check balance
go run cmd/phase_4/main.go getbalance -address "Alice"
go run cmd/phase_4/main.go getbalance -address "Bob"
```
