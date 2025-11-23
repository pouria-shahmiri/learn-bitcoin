# Multi-stage build for Bitcoin Node
# Stage 1: Builder
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the bitcoin node binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o bitcoin-node cmd/phase_11/main.go

# Build the CLI tool
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o bitcoin-cli cmd/bitcoin-cli/main.go

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates leveldb curl bash jq

# Create app directory
WORKDIR /root/

# Copy binaries from builder
COPY --from=builder /app/bitcoin-node /usr/local/bin/
COPY --from=builder /app/bitcoin-cli /usr/local/bin/
COPY --from=builder /app/scripts /usr/local/bin/scripts

# Create data directory
RUN mkdir -p /root/.bitcoin/data

# Expose ports
# 8332 - RPC port
# 8333 - P2P port
EXPOSE 8332 8333

# Set environment variables with defaults
ENV RPC_PORT=8332
ENV P2P_PORT=8333
ENV DATA_DIR=/root/.bitcoin/data
ENV NETWORK=regtest
ENV MINER_ADDRESS=""
ENV NODE_ID=node1
ENV LOG_LEVEL=info

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD bitcoin-cli --rpc-url=http://localhost:${RPC_PORT} getblockcount || exit 1

# Run the bitcoin node
CMD ["bitcoin-node"]
