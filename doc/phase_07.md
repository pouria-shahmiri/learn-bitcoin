# Phase 7: Mining & Coinbase

## 1. Objectives
Phase 7 completes the mining process. We implement the **Merkle Tree** to efficiently summarize transactions in a block header and construct the **Coinbase Transaction** to reward miners. This is where new coins are born.

## 2. Key Concepts

### Merkle Tree
A binary tree of hashes used to summarize all transactions in a block.
*   **Leaves**: Hashes of individual transactions.
*   **Root**: The "Merkle Root" is a single hash stored in the block header.
*   **Benefit**: Allows proving a transaction is in a block without downloading the whole block (SPV wallets). If one transaction changes, the root changes.

### The Coinbase Transaction
The first transaction in every block.
*   **Inputs**: None (Coinbase Input).
*   **Outputs**: Miner's address + Block Reward (e.g., 50 BTC) + Fees.
*   **Data**: Arbitrary data (e.g., "Mined by Pouria"). This is often used for signaling or just fun messages.

## 3. Code Implementation

### `MerkleTree`
Located in `pkg/utils` or `pkg/transaction`.

```go
type MerkleTree struct {
    RootNode *MerkleNode
}

func NewMerkleTree(data [][]byte) *MerkleTree {
    // Recursively hash pairs of nodes until one root remains
    // Hash(Hash(A)+Hash(B)) ...
}
```

### Mining Loop Update
The miner now performs a full setup:
1.  Pulls transactions from Mempool.
2.  Adds Coinbase transaction (Reward + Fees).
3.  Builds Merkle Tree -> Gets Root.
4.  Solves PoW (Hash < Target).

## 4. Architecture Diagram

### Merkle Tree Structure
How transactions are hashed together.

```text
       [ Merkle Root ]
             |
      +------+------+
      |             |
   [ Hash12 ]    [ Hash34 ]
      |             |
   +--+--+       +--+--+
   |     |       |     |
 [Tx1] [Tx2]   [Tx3] [Tx4]
```

### Block Structure
Where the Merkle Root fits.

```text
+-----------------------------+
|        Block Header         |
| PrevHash: 000a...           |
| MerkleRoot: ab92... <-------+
| Nonce: 12345                |
+-----------------------------+
|        Transactions         |
| 1. Coinbase (Reward)        |
| 2. Tx1                      |
| 3. Tx2                      |
+-----------------------------+
```

## 5. How to Run
Mining is now fully functional with rewards.

```bash
# Mine a block
go run cmd/phase_7/main.go mine
```
