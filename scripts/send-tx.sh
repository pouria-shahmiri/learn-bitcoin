#!/bin/bash
# send-tx.sh - Create and send test transactions

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check arguments
if [ $# -lt 3 ]; then
    echo "Usage: $0 <node_name> <to_address> <amount_sats>"
    echo ""
    echo "Available nodes: miner1, miner2, miner3, fullnode, explorer"
    echo ""
    echo "Example: $0 miner1 1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2 100000000"
    exit 1
fi

NODE_NAME=$1
TO_ADDRESS=$2
AMOUNT=$3

# Validate node name
if [[ ! "$NODE_NAME" =~ ^(miner1|miner2|miner3|fullnode|explorer)$ ]]; then
    echo -e "${RED}Error: Invalid node name${NC}"
    exit 1
fi

# Validate amount
if ! [[ "$AMOUNT" =~ ^[0-9]+$ ]]; then
    echo -e "${RED}Error: Amount must be a positive integer (in satoshis)${NC}"
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
    fullnode)
        RPC_PORT=48332
        ;;
    explorer)
        RPC_PORT=58332
        ;;
esac

echo -e "${BLUE}Creating transaction from $NODE_NAME${NC}"
echo "  To:     $TO_ADDRESS"
echo "  Amount: $AMOUNT satoshis ($(echo "scale=8; $AMOUNT/100000000" | bc) BTC)"
echo ""

# Check balance first
echo -e "${YELLOW}Checking balance...${NC}"
BALANCE=$(curl -s http://localhost:$RPC_PORT/getbalance)
echo "  Current balance: $BALANCE satoshis"

if [ "$BALANCE" -lt "$AMOUNT" ]; then
    echo -e "${RED}Error: Insufficient balance${NC}"
    exit 1
fi

# Send transaction
echo ""
echo -e "${GREEN}Sending transaction...${NC}"

# Create JSON payload
PAYLOAD=$(cat <<EOF
{
    "to_address": "$TO_ADDRESS",
    "amount": $AMOUNT
}
EOF
)

# Send via RPC
RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d "$PAYLOAD" \
    http://localhost:$RPC_PORT/sendtoaddress)

echo "Response: $RESPONSE"

# Check new balance
echo ""
echo -e "${YELLOW}Checking new balance...${NC}"
NEW_BALANCE=$(curl -s http://localhost:$RPC_PORT/getbalance)
echo "  New balance: $NEW_BALANCE satoshis"

echo ""
echo -e "${GREEN}Transaction sent successfully!${NC}"
