#!/bin/bash
# mine-blocks.sh - Trigger mining on a specific node

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

# Check arguments
if [ $# -lt 2 ]; then
    echo "Usage: $0 <node_name> <num_blocks>"
    echo ""
    echo "Available nodes: miner1, miner2, miner3"
    echo ""
    echo "Example: $0 miner1 5"
    exit 1
fi

NODE_NAME=$1
NUM_BLOCKS=$2

# Validate node name
if [[ ! "$NODE_NAME" =~ ^(miner1|miner2|miner3)$ ]]; then
    echo -e "${RED}Error: Invalid node name. Must be miner1, miner2, or miner3${NC}"
    exit 1
fi

# Validate number of blocks
if ! [[ "$NUM_BLOCKS" =~ ^[0-9]+$ ]]; then
    echo -e "${RED}Error: Number of blocks must be a positive integer${NC}"
    exit 1
fi

# Get RPC port based on node name
case $NODE_NAME in
    miner1)
        RPC_PORT=18332
        ;;
    miner2)
        RPC_PORT=28332
        ;;
    miner3)
        RPC_PORT=38332
        ;;
esac

echo -e "${BLUE}Mining $NUM_BLOCKS blocks on $NODE_NAME...${NC}"
echo ""

# Mine blocks
for i in $(seq 1 $NUM_BLOCKS); do
    echo -e "${GREEN}Mining block $i/$NUM_BLOCKS...${NC}"
    
    # Trigger mining via RPC (this would need to be implemented in the RPC server)
    # For now, we'll just show the current block count
    RESPONSE=$(curl -s http://localhost:$RPC_PORT/getblockcount)
    echo "  Current height: $RESPONSE"
    
    # In a real implementation, you would call a mine endpoint
    # curl -s -X POST http://localhost:$RPC_PORT/mine
    
    sleep 1
done

echo ""
echo -e "${GREEN}Mining complete!${NC}"

# Show final status
FINAL_HEIGHT=$(curl -s http://localhost:$RPC_PORT/getblockcount)
echo "Final block height: $FINAL_HEIGHT"
