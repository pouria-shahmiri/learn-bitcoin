#!/bin/bash
# start-testnet.sh - Launch multi-node Bitcoin testnet

set -e

echo "=== Starting Bitcoin Testnet ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "Error: Docker is not running. Please start Docker first."
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "Error: docker-compose is not installed."
    exit 1
fi

# Clean up old containers if requested
if [ "$1" == "--clean" ]; then
    echo -e "${YELLOW}Cleaning up old containers and volumes...${NC}"
    docker-compose down -v
    echo ""
fi

# Build images
echo -e "${BLUE}Building Docker images...${NC}"
docker-compose build

echo ""
echo -e "${GREEN}Starting Bitcoin network nodes...${NC}"
echo ""

# Start all nodes
docker-compose up -d

echo ""
echo -e "${GREEN}Waiting for nodes to start...${NC}"
sleep 5

# Check node status
echo ""
echo -e "${BLUE}Node Status:${NC}"
docker-compose ps

echo ""
echo -e "${GREEN}=== Bitcoin Testnet Started Successfully ===${NC}"
echo ""
echo "Available nodes:"
echo "  - Miner 1:  RPC: http://localhost:18332, P2P: localhost:18333"
echo "  - Miner 2:  RPC: http://localhost:28332, P2P: localhost:28333"
echo "  - Miner 3:  RPC: http://localhost:38332, P2P: localhost:38333"
echo "  - Full Node: RPC: http://localhost:48332, P2P: localhost:48333"
echo "  - Explorer:  RPC: http://localhost:58332, P2P: localhost:58333"
echo ""
echo "Useful commands:"
echo "  - View logs:           docker-compose logs -f [service_name]"
echo "  - View all logs:       docker-compose logs -f"
echo "  - Stop network:        docker-compose down"
echo "  - Restart node:        docker-compose restart [service_name]"
echo "  - Execute command:     docker-compose exec [service_name] bitcoin-cli [command]"
echo ""
echo "Example CLI commands:"
echo "  ./scripts/monitor.sh"
echo "  ./scripts/mine-blocks.sh miner1 5"
echo "  ./scripts/send-tx.sh miner1 [address] [amount]"
echo ""
