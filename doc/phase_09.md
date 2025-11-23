# Phase 9: Wallet & RPC

## 1. Objectives
Phase 9 focuses on usability. We build a proper **Wallet** to manage private keys and addresses securely. We also add an **RPC (Remote Procedure Call)** server so external apps (or a separate CLI tool) can talk to the node via HTTP/JSON. This separates the "Node" logic from the "User Interface".

## 2. Key Concepts

### Wallet Management
*   **Private Key**: Random 256-bit integer. Keep this secret!
*   **Public Key**: Derived from Private Key (Elliptic Curve Multiplication).
*   **Address**: Hash of Public Key (Base58Check encoded). This is what you share.
*   **Wallet File**: Stores keys on disk (`wallet.dat`), often encrypted.

### JSON-RPC
A stateless, light-weight remote procedure call (RPC) protocol.
*   **Request**: `{"method": "getbalance", "params": ["address"], "id": 1}`
*   **Response**: `{"result": 50, "error": null, "id": 1}`
*   **Why?**: Allows web apps, mobile wallets, or other scripts to interact with the blockchain without being part of the P2P network.

## 3. Code Implementation

### `Wallet` Struct
Located in `pkg/wallet`.

```go
type Wallet struct {
    PrivateKey ecdsa.PrivateKey
    PublicKey  []byte
}
```

### RPC Server
Located in `pkg/rpc`. It uses Go's `net/http` package.

```go
func StartRPCServer() {
    http.HandleFunc("/rpc", handleRPC)
    http.ListenAndServe(":8332", nil)
}
```

## 4. Architecture Diagram

### System Components
Separation of concerns.

```text
+-------------+       +-----------------+       +----------------+
|  User / UI  | ----> |   bitcoin-cli   | ----> |  Bitcoin Node  |
+-------------+       | (HTTP Client)   |       | (RPC Server)   |
                      +-----------------+       +----------------+
                                                        |
                                                        v
                                                +----------------+
                                                |  Wallet Logic  |
                                                |  (Keys/Sign)   |
                                                +----------------+
```

## 5. How to Run
Start the node with RPC enabled, then use the CLI tool to query it.

**Terminal 1 (Node):**
```bash
go run cmd/phase_9/main.go startnode
```

**Terminal 2 (CLI):**
```bash
./bitcoin-cli getbalance -address "..."
```
