# Phase 1: Basic Blockchain Structure

## 1. Objectives
The goal of Phase 1 is to establish the fundamental data structure of a blockchain. We move away from a centralized database concept to a linked list of blocks, where each block is cryptographically connected to the previous one. This ensures that history cannot be altered without breaking the chain.

## 2. Key Concepts

### What is a Blockchain?
A blockchain is a distributed public ledger. At its core, it is a **linked list** where:
1.  **Ordering**: Blocks are ordered linearly by time.
2.  **Immutability**: Each block contains the hash of the previous block. Changing data in an old block would change its hash, breaking the link to the next block and invalidating the entire subsequent chain.

### The Block
A block is a container for data. In Bitcoin, this data is transactions. In our early phase, it can be any arbitrary data (strings).

**Structure:**
*   **Timestamp**: The exact time the block was created.
*   **Data**: The actual information (e.g., "Alice sent 1 BTC to Bob").
*   **PrevBlockHash**: The SHA-256 hash of the previous block. This is the "link" in the chain.
*   **Hash**: The SHA-256 hash of the current block (calculated from all fields).

### Hashing (SHA-256)
We use **SHA-256** (Secure Hash Algorithm 256-bit). It's a one-way cryptographic function that produces a unique fixed-size string (hash) for any given input.
`Hash = SHA256(PrevBlockHash + Data + Timestamp)`

If you change even a single bit of the data, the resulting hash changes completely.

## 3. Code Implementation

### `Block` Struct
Located in `pkg/types/block.go`.

```go
type Block struct {
    Timestamp     int64
    Data          []byte
    PrevBlockHash []byte
    Hash          []byte
}
```

### `SetHash` Method
We concatenate the block fields and hash them to generate the block's unique identifier. This is a simplified version of what eventually becomes Proof-of-Work.

```go
func (b *Block) SetHash() {
    // Convert timestamp to bytes
    timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
    // Join all parts: PrevHash + Data + Timestamp
    headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
    // Calculate SHA-256
    hash := sha256.Sum256(headers)
    b.Hash = hash[:]
}
```

### `Blockchain` Struct
A simple array of blocks in memory.

```go
type Blockchain struct {
    Blocks []*Block
}
```

## 4. Architecture Diagram

### Visual Representation
Each block points to the previous one.

```text
+-------------------+        +-------------------+        +-------------------+
|    Genesis Block  |        |      Block 1      |        |      Block 2      |
+-------------------+        +-------------------+        +-------------------+
| Prev: None        | <----- | Prev: 000a...     | <----- | Prev: 004f...     |
| Data: "Genesis"   |        | Data: "Tx 1"      |        | Data: "Tx 2"      |
| Hash: 000a...     |        | Hash: 004f...     |        | Hash: 008c...     |
+-------------------+        +-------------------+        +-------------------+
```

### Flow
1.  **Genesis**: The first block is hardcoded.
2.  **Add Block**: New block takes the hash of the last block as its `PrevBlockHash`.
3.  **Link**: The chain grows linearly.

## 5. How to Run
Navigate to the project root and run the Phase 1 demo:

```bash
go run cmd/phase_1/main.go
```

**Expected Output:**
```text
Prev. hash:
Data: Genesis Block
Hash: 96a2...

Prev. hash: 96a2...
Data: Send 1 BTC to Ivan
Hash: a2f6...
...
```
