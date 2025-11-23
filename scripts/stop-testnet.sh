#!/bin/bash
# stop-testnet.sh - Stop the Bitcoin testnet

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}=== Stopping Bitcoin Testnet ===${NC}"
echo ""

# Check if containers are running
if ! docker-compose ps | grep -q "Up"; then
    echo -e "${YELLOW}No running containers found${NC}"
    exit 0
fi

# Show current status
echo -e "${YELLOW}Current containers:${NC}"
docker-compose ps
echo ""

# Ask for confirmation if not forced
if [ "$1" != "--force" ] && [ "$1" != "-f" ]; then
    read -p "Stop all nodes? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Cancelled"
        exit 0
    fi
fi

# Stop containers
echo -e "${BLUE}Stopping containers...${NC}"
docker-compose stop

echo ""
echo -e "${GREEN}All nodes stopped${NC}"
echo ""

# Ask about cleanup
if [ "$1" == "--clean" ]; then
    echo -e "${YELLOW}Removing containers and volumes...${NC}"
    docker-compose down -v
    echo -e "${GREEN}Cleanup complete${NC}"
else
    echo "To remove containers and volumes, run:"
    echo "  $0 --clean"
    echo ""
    echo "To restart the network, run:"
    echo "  ./scripts/start-testnet.sh"
fi
