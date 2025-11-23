.PHONY: all test run clean fmt vet

all: test run

# Run all tests
test:
	@echo "Running tests..."
	go test ./tests/... -v

# Run demo program (default: phase 10)
run:
	@echo "Running Milestone 10..."
	go run cmd/phase_10/main.go

# Run individual phases
phase1:
	@echo "Running Milestone 1..."
	go run cmd/phase_1/main.go

phase2:
	@echo "Running Milestone 2..."
	go run cmd/phase_2/main.go

phase3:
	@echo "Running Milestone 3..."
	go run cmd/phase_3/main.go

phase4:
	@echo "Running Milestone 4..."
	go run cmd/phase_4/main.go

phase5:
	@echo "Running Milestone 5..."
	go run cmd/phase_5/main.go

phase6:
	@echo "Running Milestone 6..."
	go run cmd/phase_6/main.go

phase7:
	@echo "Running Milestone 7..."
	go run cmd/phase_7/main.go

phase8:
	@echo "Running Milestone 8..."
	go run cmd/phase_8/main.go

phase9:
	@echo "Running Milestone 9..."
	go run cmd/phase_9/main.go

phase10:
	@echo "Running Milestone 10..."
	go run cmd/phase_10/main.go

phase11:
	@echo "Running Milestone 11..."
	go run cmd/phase_11/main.go

# Docker commands
docker-build:
	@echo "Building Docker image..."
	docker build -t bitcoin-node:latest .

docker-up:
	@echo "Starting Docker network..."
	./scripts/start-testnet.sh

docker-down:
	@echo "Stopping Docker network..."
	./scripts/stop-testnet.sh

docker-clean:
	@echo "Cleaning Docker network..."
	./scripts/stop-testnet.sh --clean

docker-logs:
	@echo "Showing Docker logs..."
	docker-compose logs -f

docker-monitor:
	@echo "Monitoring network..."
	./scripts/monitor.sh

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Clean build artifacts
clean:
	go clean
	rm -f coverage.out bitcoin-node bitcoin-cli

# Run tests with coverage
coverage:
	go test ./tests/... -coverprofile=coverage.out
	go tool cover -html=coverage.out

# Check for common mistakes
check: fmt vet test

# Install dependencies
deps:
	go mod download
	go mod tidy