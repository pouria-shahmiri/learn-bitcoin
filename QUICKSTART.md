# Phase 11: Quick Start Guide

Get your multi-node Bitcoin network running in 5 minutes!

## Prerequisites

- Docker installed and running
- docker-compose installed
- At least 2GB free RAM
- Ports 18332-58333 available

## Quick Start (3 Commands)

```bash
# 1. Start the network
./scripts/start-testnet.sh

# 2. Monitor the network
./scripts/monitor.sh

# 3. Stop the network (when done)
./scripts/stop-testnet.sh
```

## Interactive Demo

For a guided walkthrough of all features:

```bash
./scripts/demo.sh
```

This will:
- ✓ Check prerequisites
- ✓ Build Docker images
- ✓ Start 5 nodes
- ✓ Show mining activity
- ✓ Demonstrate RPC endpoints
- ✓ Monitor network sync

## What You Get

### 5 Running Nodes

| Node | Type | RPC Port | Auto-Mine | Interval |
|------|------|----------|-----------|----------|
| miner1 | Mining | 18332 | Yes | 10s |
| miner2 | Mining | 28332 | Yes | 15s |
| miner3 | Mining | 38332 | Yes | 20s |
| fullnode | Full Node | 48332 | No | - |
| explorer | Monitor | 58332 | No | - |

### Connected Network

All nodes are connected via P2P and sync blocks automatically.

### Persistent Storage

Each node has its own Docker volume for blockchain data.

## Common Commands

### View Logs

```bash
# All nodes
docker-compose logs -f

# Specific node
docker-compose logs -f miner1

# Last 50 lines
docker-compose logs --tail=50 miner1
```

### Check Status

```bash
# Quick status
./scripts/monitor.sh

# Block count
curl http://localhost:18332/getblockcount

# Balance
curl http://localhost:18332/getbalance

# Container status
docker-compose ps
```

### Execute Commands

```bash
# Get shell in container
docker-compose exec miner1 /bin/sh

# Run CLI command
docker-compose exec miner1 bitcoin-cli getblockcount
```

### Restart Nodes

```bash
# Restart specific node
docker-compose restart miner1

# Restart all nodes
docker-compose restart
```

## Testing Scenarios

### Watch Mining

```bash
# Watch miner1 logs
docker-compose logs -f miner1 | grep "Mined block"
```

### Check Sync

```bash
# Run monitor every 5 seconds
watch -n 5 ./scripts/monitor.sh
```

### Test Network Partition

```bash
# Disconnect miner3
docker network disconnect learn-bitcoin_bitcoin_net bitcoin-miner3

# Wait 30 seconds
sleep 30

# Reconnect
docker network connect learn-bitcoin_bitcoin_net bitcoin-miner3

# Watch reorg
docker-compose logs -f miner3
```

## Troubleshooting

### Containers won't start

```bash
# Check Docker
docker info

# Check ports
netstat -tulpn | grep -E '(18332|28332|38332)'

# Clean restart
./scripts/start-testnet.sh --clean
```

### Nodes not syncing

```bash
# Check connectivity
docker-compose exec miner1 ping miner2

# Restart
docker-compose restart
```

### Out of space

```bash
# Check usage
docker system df

# Clean
docker-compose down -v
docker system prune -a
```

## Next Steps

1. **Explore RPC API**
   ```bash
   curl http://localhost:18332/getblockcount
   curl http://localhost:18332/getbalance
   curl http://localhost:18332/listaddresses
   ```

2. **Watch Network Activity**
   ```bash
   docker-compose logs -f
   ```

3. **Test Transactions**
   ```bash
   ./scripts/send-tx.sh miner1 <address> <amount>
   ```

4. **Monitor Continuously**
   ```bash
   watch -n 5 ./scripts/monitor.sh
   ```

## Clean Up

```bash
# Stop (keep data)
./scripts/stop-testnet.sh

# Stop and remove data
./scripts/stop-testnet.sh --clean

# Or manually
docker-compose down -v
```

## Help

For detailed documentation, see:
- [Phase 11 README](../docs/PHASE_11_README.md)
- [Configuration Template](../config.template)

For issues:
1. Check logs: `docker-compose logs`
2. Check status: `docker-compose ps`
3. Try clean restart: `./scripts/start-testnet.sh --clean`

## Success Indicators

You know it's working when:

✓ All 5 containers show "Up" status
✓ Miners are producing blocks (check logs)
✓ All nodes have the same block height
✓ RPC endpoints respond
✓ Peer connections are established

Check with:
```bash
./scripts/monitor.sh
```

Happy mining! 🚀
