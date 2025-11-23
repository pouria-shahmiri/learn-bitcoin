# Phase 6: Mempool & Fee Policy

## 1. Objectives
In a real network, transactions aren't mined instantly. They wait in a **Memory Pool (Mempool)**. Miners select transactions from the mempool to include in the next block, prioritizing those with higher fees. This phase introduces the concept of unconfirmed transactions.

## 2. Key Concepts

### The Mempool
A temporary storage (in RAM) for valid transactions that have not yet been included in a block.
*   **Validation**: Before entering the mempool, a transaction is checked (signatures, available funds).
*   **Life Cycle**: 
    1.  **Created**: User sends coins.
    2.  **Mempool**: Transaction waits here.
    3.  **Block**: Miner picks it up.
    4.  **Removed**: Once in a block, it's removed from the mempool.

### Transaction Fees
Miners need an incentive to include your transaction.
*   **Formula**: `Fee = Input Sum - Output Sum`.
*   **Example**: Input (50 BTC) - Output (49 BTC) = 1 BTC Fee.
*   **Collection**: The miner claims this fee in their Coinbase transaction.

## 3. Code Implementation

### `Mempool` Struct
Located in `pkg/mempool`.

```go
type Mempool struct {
    Transactions map[string]*Transaction
}
```

### Fee Calculation
When selecting transactions for a block, the miner sorts them.

```go
func (m *Mempool) SelectTransactions(maxBlockSize int) []*Transaction {
    // 1. Calculate fee for each tx
    // 2. Sort by Fee (or Fee/Size ratio)
    // 3. Pick top transactions until block is full
    // 4. Return selected list
}
```

## 4. Architecture Diagram

### Transaction Lifecycle
From creation to confirmation.

```text
User
 |
 v
[ Create Tx ]
 |
 v
[ Mempool ] <----------------+
 |  (Waiting Room)           |
 |                           |
 +---(Select High Fee)--> [ Miner ]
                             |
                             v
                        [ New Block ]
                             |
                             v
                        [ Blockchain ]
```

## 5. How to Run
Phase 6 usually introduces a persistent server or a long-running process, but can still be tested via CLI if designed that way.

```bash
# Start a node (if applicable) or use CLI to simulate
go run cmd/phase_6/main.go startnode
```
*(Note: Specific commands depend on the implementation details of this phase)*
