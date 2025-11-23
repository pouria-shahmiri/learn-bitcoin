# Learn Bitcoin: A Go Implementation

**Build your own Bitcoin node from scratch in Go.**

This project is a comprehensive, step-by-step guide to understanding the inner workings of Bitcoin. By building it yourself, you will learn about blockchain architecture, Proof-of-Work consensus, P2P networking, cryptography, and distributed systems.

The project is divided into **11 Phases**, each introducing a core concept and adding complexity to the system.

## 📚 Documentation & Roadmap

Full documentation for each phase is available in the [`doc/`](./doc) directory.

| Phase | Topic | Description |
| :--- | :--- | :--- |
| **[Phase 1](./doc/phase_01.md)** | **Basic Blockchain** | The fundamental linked-list structure of blocks and hashing. |
| **[Phase 2](./doc/phase_02.md)** | **Proof of Work** | Implementing the mining algorithm to secure the chain. |
| **[Phase 3](./doc/phase_03.md)** | **Persistence & CLI** | Saving data to disk (LevelDB) and interacting via command line. |
| **[Phase 4](./doc/phase_04.md)** | **Transactions** | Moving from string data to real input/output transactions. |
| **[Phase 5](./doc/phase_05.md)** | **UTXO Set** | Optimizing validation with a Unspent Transaction Output set. |
| **[Phase 6](./doc/phase_06.md)** | **Mempool** | Handling unconfirmed transactions and fee prioritization. |
| **[Phase 7](./doc/phase_07.md)** | **Mining & Merkle** | Constructing blocks with Merkle Trees and Coinbase rewards. |
| **[Phase 8](./doc/phase_08.md)** | **P2P Network** | Nodes discovering and syncing with each other over TCP. |
| **[Phase 9](./doc/phase_09.md)** | **Wallet & RPC** | Managing keys/addresses and exposing an API. |
| **[Phase 10](./doc/phase_10.md)** | **Consensus** | Handling forks, reorgs, and orphan blocks. |
| **[Phase 11](./doc/phase_11.md)** | **Deployment** | Dockerizing the network for easy deployment. |

## 🏗️ Architecture Overview

The system mimics the architecture of Bitcoin Core.

```text
+---------------------------------------------------------------+
|                        CLI / RPC Client                       |
+-------------------------------+-------------------------------+
                                | HTTP / JSON-RPC
+-------------------------------v-------------------------------+
|                        Bitcoin Node                           |
|                                                               |
|  +-----------+      +-----------+      +-------------------+  |
|  |  Wallet   |      |  Mempool  |      |    Miner (PoW)    |  |
|  +-----------+      +-----------+      +-------------------+  |
|        ^                  ^                      |            |
|        |                  |                      v            |
|  +---------------------------------------------------------+  |
|  |                   Blockchain Core                       |  |
|  |      (Validation, Reorgs, UTXO Set, Merkle Trees)       |  |
|  +---------------------------------------------------------+  |
|        ^                                         ^            |
|        | Read/Write                              | P2P        |
|  +-----v------+                          +-------v-------+    |
|  |  LevelDB   |                          |  P2P Network  |    |
|  +------------+                          +---------------+    |
+---------------------------------------------------------------+
```

## 🚀 Quick Start

You can run any phase individually to see the progress.

### Prerequisites
*   **Go** 1.22 or higher
*   **Make** (optional, for helper commands)

### Running Phase 1 (Basic Chain)
```bash
go run cmd/phase_1/main.go
```

### Running Phase 11 (Full Docker Network)
```bash
docker-compose up --build
```

### Automated Transactions

A transaction generator is included to simulate network activity.

**Using Docker:**
The `tx-sender` service automatically starts with the network and sends a transaction every 3 seconds from `miner1` to `miner2`.

View logs:
```bash
docker-compose logs -f tx-sender
```

**Manual Usage:**
You can also run the script manually against any running node:

```bash
# Send from miner1 to a specific address every 5 seconds
export TARGET_URL="http://localhost:18332"
export TO_ADDRESS="1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2"
export INTERVAL=5
./scripts/auto-tx.sh
```

### Helper Scripts

The `scripts/` directory contains several useful tools for managing the testnet:

| Script | Description | Usage |
| :--- | :--- | :--- |
| `start-testnet.sh` | Starts the entire Docker network | `./scripts/start-testnet.sh [--clean]` |
| `stop-testnet.sh` | Stops the network | `./scripts/stop-testnet.sh [--clean]` |
| `monitor.sh` | Shows status of all nodes | `./scripts/monitor.sh` |
| `mine-blocks.sh` | Manually triggers mining | `./scripts/mine-blocks.sh <node> <count>` |
| `send-tx.sh` | Sends a single transaction | `./scripts/send-tx.sh <node> <addr> <amt>` |
| `demo.sh` | Runs a full interactive demo | `./scripts/demo.sh` |

## 🤝 Contributing
Feel free to open issues or submit PRs if you find bugs or want to improve the explanations. This is a learning project!

## 📜 License
MIT
