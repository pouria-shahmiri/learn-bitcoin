# Phase 11: Docker Deployment

## 1. Objectives
The final phase moves from running nodes in local terminals to a production-like environment using **Docker**. We containerize the application and use **Docker Compose** to orchestrate a multi-node network with a single command. This ensures consistency across different environments.

## 2. Key Concepts

### Dockerfile
A recipe for building the image of our node.
*   **Base Image**: `golang:1.22-alpine` (Lightweight Linux).
*   **Build Stage**: Compiles the binary.
*   **Run Stage**: Minimal image with just the binary to keep size down.

### Docker Compose
Defines services, networks, and volumes.
*   **Services**: `miner-1`, `miner-2`, `wallet-node`.
*   **Network**: A private bridge network (`bitcoin-net`) allowing nodes to talk by hostname (e.g., `miner-1:3000`).
*   **Volumes**: Persistent storage for blockchain data so it survives container restarts.

## 3. Code Implementation

### `Dockerfile`
Located in project root.

```dockerfile
# Stage 1: Build
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o bitcoind ./cmd/phase_11

# Stage 2: Run
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/bitcoind .
CMD ["./bitcoind"]
```

### `docker-compose.yml`
Defines the cluster.

```yaml
services:
  node1:
    build: .
    environment:
      - NODE_ID=3000
    ports:
      - "3000:3000"
```

## 4. Architecture Diagram

### Container Network
Isolated environment.

```text
Host Machine
|
+--- [ Docker Engine ] -------------------------+
|                                               |
|  +------------------+    +------------------+ |
|  | Container: Node1 |    | Container: Node2 | |
|  | IP: 172.18.0.2   |<-->| IP: 172.18.0.3   | |
|  +------------------+    +------------------+ |
|           ^                       ^           |
|           | (Volume Mount)        |           |
|      /tmp/node1_data         /tmp/node2_data  |
|                                               |
+-----------------------------------------------+
```

## 5. How to Run
Deploy the entire network with one command.

```bash
# Build and start
docker-compose up --build

# Check logs
docker-compose logs -f

# Stop
docker-compose down
```
