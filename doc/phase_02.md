# Phase 2: Proof of Work

## 1. Objectives
In Phase 1, adding blocks was instant and cheap. This is dangerous for a public blockchain because anyone could spam the network or rewrite history effortlessly. Phase 2 introduces **Proof of Work (PoW)**, a mechanism that makes adding blocks computationally expensive, securing the network against attacks.

## 2. Key Concepts

### The Problem: Spam and Rewrites
If creating a block is free, an attacker can:
1.  **Spam**: Fill the chain with garbage data, bloating the ledger size.
2.  **Rewrite**: Change an old block and re-calculate all subsequent hashes instantly, effectively altering history (e.g., removing a transaction where they paid you).

### The Solution: Proof of Work
PoW requires the miner to solve a difficult mathematical puzzle before a block is accepted.
*   **Puzzle**: Find a number (Nonce) such that `Hash(Block Data + Nonce) < Target`.
*   **Target**: A threshold defined by the network difficulty. The lower the target, the harder it is to find a hash below it.
*   **Nonce**: "Number used once". A counter the miner increments to change the block's hash.

### Difficulty
The difficulty adjusts over time to keep the block generation rate constant (e.g., 10 minutes in Bitcoin).
*   **High Difficulty**: Target is small (starts with many zeros). Hard to find.
*   **Low Difficulty**: Target is large. Easy to find.

## 3. Code Implementation

### `ProofOfWork` Struct
Located in `pkg/mining`. It holds the block we want to mine and the target difficulty.

```go
type ProofOfWork struct {
    block  *Block
    target *big.Int
}
```

### The Mining Loop (`Run`)
This is the core of the miner. It loops millions of times until it finds a valid hash.

```go
func (pow *ProofOfWork) Run() (int, []byte) {
    var hashInt big.Int
    var hash [32]byte
    nonce := 0

    for nonce < maxNonce {
        // 1. Prepare data: PrevHash + Data + Timestamp + NONCE
        data := pow.prepareData(nonce) 
        hash = sha256.Sum256(data)
        hashInt.SetBytes(hash[:])

        // 2. Check if Hash < Target
        if hashInt.Cmp(pow.target) == -1 {
            break // Success! We found a valid nonce.
        } else {
            nonce++ // Try next number
        }
    }
    return nonce, hash[:]
}
```

## 4. Architecture Diagram

### Mining Process
The miner acts like a lottery player, trying random numbers (nonces) until they win.

```text
+---------+    +---------+    +---------+
| Block   | +  | Nonce 0 | -> | Hash... | > Target? (Fail)
+---------+    +---------+    +---------+

+---------+    +---------+    +---------+
| Block   | +  | Nonce 1 | -> | Hash... | > Target? (Fail)
+---------+    +---------+    +---------+
     ...            ...            ...
+---------+    +---------+    +---------+
| Block   | +  | Nonce X | -> | 0000... | < Target? (SUCCESS!)
+---------+    +---------+    +---------+
                                   |
                                   v
                            Add to Blockchain
```

## 5. How to Run
Run the Phase 2 demo to see mining in action. You will notice a delay as the computer "works" to find the hash.

```bash
go run cmd/phase_2/main.go
```

**Expected Output:**
```text
Mining the block containing "Genesis Block"
0000x8s7... (Valid Hash found!)

Mining the block containing "Data 1"
0000a7d2... (Valid Hash found!)
```
