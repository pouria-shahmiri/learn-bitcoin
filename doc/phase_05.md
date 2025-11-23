# Phase 5: UTXO Set & Validation

## 1. Objectives
As the blockchain grows, scanning the entire history (every block from Genesis) to calculate a balance becomes incredibly slow. Phase 5 introduces the **UTXO Set**, a separate database that caches only the unspent outputs. We also implement stricter chain validation to ensure no double-spending occurs.

## 2. Key Concepts

### The UTXO Set
A subset of the blockchain data that only contains outputs that haven't been spent yet.
*   **Purpose**: Fast transaction validation and balance calculation.
*   **Size**: Much smaller than the full blockchain.
*   **Update Mechanism**: When a new block is added:
    1.  **Remove** outputs that were spent by inputs in the new block.
    2.  **Add** new outputs created by the new block.

### Double Spend Protection
We must ensure an output is not spent twice.
*   **Check**: Before accepting a transaction, we verify that its inputs refer to outputs that exist in the UTXO set.
*   **Result**: If the output is not in the set, it's already spent or never existed.

## 3. Code Implementation

### `UTXOSet` Struct
Located in `pkg/utxo/utxo_set.go`. It interacts with the database to store the set.

```go
type UTXOSet struct {
    Blockchain *Blockchain
}

func (u *UTXOSet) Reindex() {
    // 1. Wipe existing UTXO bucket in DB
    // 2. Scan full chain from Genesis
    // 3. Identify unspent outputs
    // 4. Save them to DB
}
```

### `FindSpendableOutputs`
Now uses the UTXO set instead of the full chain, making `send` commands much faster (O(N) where N is unspent outputs, vs O(Total History)).

```go
func (u *UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
    // Iterate only through UTXO database
    // Accumulate values until amount is reached
}
```

## 4. Architecture Diagram

### UTXO Set Logic
The UTXO set acts as a fast cache layer.

```text
       Blockchain (Full History)
+-------+    +-------+    +-------+
| Blk 1 | -> | Blk 2 | -> | Blk 3 |
+-------+    +-------+    +-------+
    |            |            |
    v            v            v
  (Scan)       (Scan)       (Scan)
    |            |            |
    +------------+------------+
                 |
                 v
        +------------------+
        |     UTXO Set     |
        | (Unspent Only)   |
        +------------------+
                 ^
                 |
           Get Balance()
             (Fast!)
```

## 5. How to Run
You might need to reindex your chain if you are upgrading from Phase 4.

```bash
# Rebuild the UTXO set index
go run cmd/phase_5/main.go reindexutxo

# Check balance (Should be faster)
go run cmd/phase_5/main.go getbalance -address "Alice"
```
