#!/bin/bash
# demo.sh - Complete Phase 11 demonstration

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}"
cat << "EOF"
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║     Bitcoin Node - Phase 11 Demonstration                ║
║     Docker Deployment & Multi-Node Orchestration         ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
EOF
echo -e "${NC}"

# Function to print section header
section() {
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}  $1${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
    echo ""
}

# Function to wait for user
wait_user() {
    echo ""
    read -p "Press Enter to continue..."
    echo ""
}

# Check prerequisites
section "1. Checking Prerequisites"

echo -e "${YELLOW}Checking Docker...${NC}"
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}Error: Docker is not running${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Docker is running${NC}"

echo -e "${YELLOW}Checking docker-compose...${NC}"
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}Error: docker-compose is not installed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ docker-compose is installed${NC}"

wait_user

# Clean previous runs
section "2. Cleaning Previous Runs"

echo -e "${YELLOW}Stopping any running containers...${NC}"
docker-compose down -v 2>/dev/null || true
echo -e "${GREEN}✓ Cleanup complete${NC}"

wait_user

# Build Docker image
section "3. Building Docker Image"

echo -e "${YELLOW}Building bitcoin-node Docker image...${NC}"
echo "This may take a few minutes on first run..."
docker-compose build
echo -e "${GREEN}✓ Docker image built successfully${NC}"

wait_user

# Start the network
section "4. Starting Multi-Node Network"

echo -e "${YELLOW}Starting 5 Bitcoin nodes:${NC}"
echo "  - 3 Mining nodes (miner1, miner2, miner3)"
echo "  - 1 Full node (fullnode)"
echo "  - 1 Explorer node (explorer)"
echo ""

docker-compose up -d

echo -e "${GREEN}✓ All nodes started${NC}"
echo ""
echo "Waiting for nodes to initialize..."
sleep 10

wait_user

# Show node status
section "5. Node Status"

echo -e "${YELLOW}Container Status:${NC}"
docker-compose ps
echo ""

echo -e "${YELLOW}Node Information:${NC}"
echo ""

for node in miner1 miner2 miner3 fullnode explorer; do
    case $node in
        miner1) port=18332 ;;
        miner2) port=28332 ;;
        miner3) port=38332 ;;
        fullnode) port=48332 ;;
        explorer) port=58332 ;;
    esac
    
    echo -e "${CYAN}$node (RPC: $port)${NC}"
    height=$(curl -s http://localhost:$port/getblockcount 2>/dev/null || echo "N/A")
    balance=$(curl -s http://localhost:$port/getbalance 2>/dev/null || echo "N/A")
    echo "  Height:  $height"
    echo "  Balance: $balance sats"
    echo ""
done

wait_user

# Show logs
section "6. Viewing Node Logs"

echo -e "${YELLOW}Recent activity from miners:${NC}"
echo ""

for node in miner1 miner2 miner3; do
    echo -e "${CYAN}=== $node ===${NC}"
    docker-compose logs --tail=10 $node | grep -E "INFO|Mined|Connected|Status" || echo "No activity yet"
    echo ""
done

wait_user

# Test RPC endpoints
section "7. Testing RPC Endpoints"

echo -e "${YELLOW}Testing miner1 RPC endpoints:${NC}"
echo ""

echo -e "${CYAN}GET /getblockcount${NC}"
curl -s http://localhost:18332/getblockcount
echo -e "\n"

echo -e "${CYAN}GET /getbalance${NC}"
curl -s http://localhost:18332/getbalance
echo -e "\n"

echo -e "${CYAN}GET /listaddresses${NC}"
curl -s http://localhost:18332/listaddresses
echo -e "\n"

wait_user

# Monitor network
section "8. Network Monitoring"

echo -e "${YELLOW}Running network monitor...${NC}"
./scripts/monitor.sh

wait_user

# Show peer connections
section "9. Peer Connections"

echo -e "${YELLOW}Checking peer connections:${NC}"
echo ""

for node in miner1 miner2 miner3; do
    echo -e "${CYAN}$node connections:${NC}"
    docker-compose logs $node | grep "Connected to peer" | tail -3 || echo "  No connections logged yet"
    echo ""
done

wait_user

# Wait for mining
section "10. Watching Mining Activity"

echo -e "${YELLOW}Watching mining activity for 30 seconds...${NC}"
echo "You should see blocks being mined by different nodes"
echo ""

timeout 30 docker-compose logs -f miner1 miner2 miner3 | grep "Mined block" || true

echo ""
echo -e "${GREEN}Mining activity captured${NC}"

wait_user

# Final status
section "11. Final Network Status"

./scripts/monitor.sh

wait_user

# Cleanup options
section "12. Demo Complete"

echo -e "${GREEN}Phase 11 demonstration complete!${NC}"
echo ""
echo "The network is still running. You can:"
echo ""
echo "  1. View logs:          docker-compose logs -f"
echo "  2. Monitor network:    ./scripts/monitor.sh"
echo "  3. Test RPC:           curl http://localhost:18332/getblockcount"
echo "  4. Execute commands:   docker-compose exec miner1 bitcoin-cli getbalance"
echo ""
echo "To stop the network:"
echo "  ./scripts/stop-testnet.sh"
echo ""
echo "To stop and clean all data:"
echo "  ./scripts/stop-testnet.sh --clean"
echo ""

read -p "Do you want to stop the network now? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    echo -e "${YELLOW}Stopping network...${NC}"
    ./scripts/stop-testnet.sh --force
    echo -e "${GREEN}Network stopped${NC}"
else
    echo ""
    echo -e "${GREEN}Network is still running. Happy testing!${NC}"
fi

echo ""
echo -e "${BLUE}Thank you for trying Phase 11!${NC}"
