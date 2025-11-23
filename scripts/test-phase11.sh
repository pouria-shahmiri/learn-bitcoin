#!/bin/bash
# test-phase11.sh - Automated tests for Phase 11

# set -e removed to allow all tests to run

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

TESTS_PASSED=0
TESTS_FAILED=0

# Test function
test_case() {
    local name=$1
    local command=$2
    local expected=$3
    
    echo -n "Testing: $name... "
    
    if result=$(eval "$command" 2>&1); then
        if [ -z "$expected" ] || echo "$result" | grep -q "$expected"; then
            echo -e "${GREEN}PASS${NC}"
            ((TESTS_PASSED++))
            return 0
        fi
    fi
    
    echo -e "${RED}FAIL${NC}"
    echo "  Command: $command"
    echo "  Result: $result"
    ((TESTS_FAILED++))
    return 1
}

echo "=== Phase 11 Automated Tests ==="
echo ""

# Check if Docker is available
DOCKER_AVAILABLE=false
if docker info > /dev/null 2>&1; then
    DOCKER_AVAILABLE=true
    echo -e "${GREEN}Docker is available${NC}"
else
    echo -e "${YELLOW}Docker not available - skipping Docker tests${NC}"
fi
echo ""

# Test 1: docker-compose is available (only if Docker is available)
if [ "$DOCKER_AVAILABLE" = true ]; then
    test_case "docker-compose is available" "command -v docker-compose" ""
fi


# Test 3: Build succeeds
echo ""
echo "Building project..."
test_case "Go build succeeds" "go build -o bitcoin-node cmd/phase_11/main.go" ""

# Test 4: Scripts are executable
test_case "start-testnet.sh is executable" "test -x scripts/start-testnet.sh" ""
test_case "stop-testnet.sh is executable" "test -x scripts/stop-testnet.sh" ""
test_case "monitor.sh is executable" "test -x scripts/monitor.sh" ""
test_case "demo.sh is executable" "test -x scripts/demo.sh" ""

# Test 5: Configuration files exist
test_case "Dockerfile exists" "test -f Dockerfile" ""
test_case "docker-compose.yml exists" "test -f docker-compose.yml" ""
test_case "config.template exists" "test -f config.template" ""

# Test 6: Required packages exist
test_case "config package exists" "test -f pkg/config/config.go" ""
test_case "network/server.go exists" "test -f pkg/network/server.go" ""

# Test 7: Documentation exists
test_case "Phase 11 README exists" "test -f docs/PHASE_11_README.md" ""
test_case "QUICKSTART.md exists" "test -f QUICKSTART.md" ""

echo ""
echo "=== Test Summary ==="
echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Failed: ${RED}$TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    echo ""
    echo "Ready to run:"
    echo "  ./scripts/start-testnet.sh"
    echo "  ./scripts/demo.sh"
    exit 0
else
    echo -e "${RED}Some tests failed${NC}"
    exit 1
fi
