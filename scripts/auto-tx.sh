#!/bin/bash
set -e

# Default values
TARGET_URL=${TARGET_URL:-"http://miner1:8332"}
TO_ADDRESS=${TO_ADDRESS:-"1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2"} # Default to miner2 address
AMOUNT=${AMOUNT:-100000} # 0.001 BTC
INTERVAL=${INTERVAL:-3}

echo "Starting auto-tx sender..."
echo "Target: $TARGET_URL"
echo "To: $TO_ADDRESS"
echo "Amount: $AMOUNT"
echo "Interval: $INTERVAL seconds"

# Wait for target node to be ready
echo "Waiting for target node to be ready..."
until curl -s "$TARGET_URL/getblockcount" > /dev/null; do
  echo "Target node not ready, waiting..."
  sleep 5
done

while true; do
    # Check balance
    BALANCE_RESP=$(curl -s "$TARGET_URL/getbalance")
    BALANCE=$(echo "$BALANCE_RESP" | jq -r '.result.balance // 0')
    
    echo "Current balance: $BALANCE"
    
    if [ "$BALANCE" -lt "$AMOUNT" ]; then
        echo "Insufficient funds ($BALANCE < $AMOUNT). Waiting for mining..."
        sleep 5
        continue
    fi

    echo "Sending transaction..."
    
    # Construct JSON payload
    PAYLOAD="{\"address\": \"$TO_ADDRESS\", \"amount\": $AMOUNT}"
    
    # Send request
    RESPONSE=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$PAYLOAD" \
        "$TARGET_URL/sendtoaddress")
        
    echo "Response: $RESPONSE"
    
    sleep $INTERVAL
done
