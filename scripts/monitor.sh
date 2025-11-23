#!/bin/bash
# monitor.sh - Monitor network status across all nodes

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

# Function to get node info
get_node_info() {
    local node_name=$1
    local rpc_port=$2
    
    # Get block count
    local height=$(curl -s http://localhost:$rpc_port/getblockcount 2>/dev/null || echo "N/A")
    
    # Get balance
    local balance=$(curl -s http://localhost:$rpc_port/getbalance 2>/dev/null || echo "N/A")
    
    # Get peer count (if endpoint exists)
    # local peers=$(curl -s http://localhost:$rpc_port/getpeerinfo 2>/dev/null | jq length || echo "N/A")
    local peers="N/A"
    
    echo -e "${CYAN}$node_name${NC} (port $rpc_port)"
    echo "  Height:  $height"
    echo "  Balance: $balance sats"
    echo "  Peers:   $peers"
    echo ""
}

# Clear screen
clear

echo -e "${BLUE}=== Bitcoin Network Monitor ===${NC}"
echo ""
echo "Timestamp: $(date)"
echo ""

# Check if containers are running
echo -e "${YELLOW}Container Status:${NC}"
docker-compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || echo "No containers running"
echo ""

# Monitor each node
echo -e "${YELLOW}Node Information:${NC}"
echo ""

get_node_info "Miner 1" 18332
get_node_info "Miner 2" 28332
get_node_info "Miner 3" 38332
get_node_info "Full Node" 48332
get_node_info "Explorer" 58332

# Show recent logs from all miners
echo -e "${YELLOW}Recent Activity (last 5 lines from each miner):${NC}"
echo ""

for node in miner1 miner2 miner3; do
    echo -e "${CYAN}=== $node ===${NC}"
    docker-compose logs --tail=5 $node 2>/dev/null | grep -E "Mined block|Connected to peer|Status" || echo "No recent activity"
    echo ""
done

# Network statistics
echo -e "${YELLOW}Network Statistics:${NC}"
echo ""

# Get heights from all nodes
HEIGHT1=$(curl -s http://localhost:18332/getblockcount 2>/dev/null || echo "0")
HEIGHT2=$(curl -s http://localhost:28332/getblockcount 2>/dev/null || echo "0")
HEIGHT3=$(curl -s http://localhost:38332/getblockcount 2>/dev/null || echo "0")
HEIGHT4=$(curl -s http://localhost:48332/getblockcount 2>/dev/null || echo "0")
HEIGHT5=$(curl -s http://localhost:58332/getblockcount 2>/dev/null || echo "0")

# Calculate max height
MAX_HEIGHT=$HEIGHT1
[ "$HEIGHT2" -gt "$MAX_HEIGHT" ] && MAX_HEIGHT=$HEIGHT2
[ "$HEIGHT3" -gt "$MAX_HEIGHT" ] && MAX_HEIGHT=$HEIGHT3
[ "$HEIGHT4" -gt "$MAX_HEIGHT" ] && MAX_HEIGHT=$HEIGHT4
[ "$HEIGHT5" -gt "$MAX_HEIGHT" ] && MAX_HEIGHT=$HEIGHT5

echo "  Network Height: $MAX_HEIGHT"
echo "  Sync Status:"
echo "    Miner 1:   $HEIGHT1 $([ "$HEIGHT1" == "$MAX_HEIGHT" ] && echo -e "${GREEN}✓${NC}" || echo -e "${YELLOW}syncing...${NC}")"
echo "    Miner 2:   $HEIGHT2 $([ "$HEIGHT2" == "$MAX_HEIGHT" ] && echo -e "${GREEN}✓${NC}" || echo -e "${YELLOW}syncing...${NC}")"
echo "    Miner 3:   $HEIGHT3 $([ "$HEIGHT3" == "$MAX_HEIGHT" ] && echo -e "${GREEN}✓${NC}" || echo -e "${YELLOW}syncing...${NC}")"
echo "    Full Node: $HEIGHT4 $([ "$HEIGHT4" == "$MAX_HEIGHT" ] && echo -e "${GREEN}✓${NC}" || echo -e "${YELLOW}syncing...${NC}")"
echo "    Explorer:  $HEIGHT5 $([ "$HEIGHT5" == "$MAX_HEIGHT" ] && echo -e "${GREEN}✓${NC}" || echo -e "${YELLOW}syncing...${NC}")"

echo ""
echo -e "${GREEN}Monitor complete!${NC}"
echo ""
echo "To watch logs in real-time:"
echo "  docker-compose logs -f"
echo ""
echo "To run this monitor continuously:"
echo "  watch -n 5 ./scripts/monitor.sh"
