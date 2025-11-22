.PHONY: all test run clean fmt vet

all: test run

# Run all tests
test:
	@echo "Running tests..."
	go test ./tests/... -v

# Run demo program
run:
	@echo "Running Milestone 2..."
	go run cmd/phase_2/main.go

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