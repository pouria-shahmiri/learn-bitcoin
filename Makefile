.PHONY: all test run clean fmt vet

all: test run

# Run all tests
test:
	@echo "Running tests..."
	go test ./tests/... -v

# Run demo program (default: phase 6)
run:
	@echo "Running Milestone 6..."
	go run cmd/phase_6/main.go

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

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Clean build artifacts
clean:
	go clean
	rm -f coverage.out

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