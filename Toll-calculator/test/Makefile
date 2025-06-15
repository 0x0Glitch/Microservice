# Test Makefile for Toll Calculator
.PHONY: test test-unit test-integration test-all test-coverage test-benchmark test-race clean help

# Default target
help:
	@echo "Available test targets:"
	@echo "  test           - Run unit tests only"
	@echo "  test-unit      - Run unit tests with verbose output"
	@echo "  test-integration - Run integration tests (requires services running)"
	@echo "  test-all       - Run all tests (unit + integration)"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-benchmark - Run benchmark tests"
	@echo "  test-race      - Run tests with race detection"
	@echo "  test-short     - Run tests in short mode (skip long-running tests)"
	@echo "  clean          - Clean test cache and coverage files"

# Basic unit tests
test:
	@echo "Running unit tests..."
	cd .. && go test -v ./test/unit/...

# Unit tests with verbose output
test-unit:
	@echo "Running unit tests with verbose output..."
	cd .. && go test -v -count=1 ./test/unit/...

# Integration tests (requires services to be running)
test-integration:
	@echo "Running integration tests..."
	@echo "Note: Make sure aggregator service is running on localhost:3000"
	cd .. && go test -v -count=1 ./test/integration/...

# Run all tests
test-all:
	@echo "Running all tests..."
	cd .. && go test -v -count=1 ./test/...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	cd .. && go test -v -coverprofile=coverage.out -covermode=atomic ./test/unit/...
	cd .. && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmark tests
test-benchmark:
	@echo "Running benchmark tests..."
	cd .. && go test -v -bench=. -benchmem ./test/unit/...

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	cd .. && go test -v -race -count=1 ./test/unit/...

# Run tests in short mode
test-short:
	@echo "Running tests in short mode..."
	cd .. && go test -v -short ./test/...

# Run specific test
test-specific:
	@echo "Usage: make test-specific TEST=TestName"
	@if [ -z "$(TEST)" ]; then echo "Please specify TEST=TestName"; exit 1; fi
	cd .. && go test -v -run $(TEST) ./test/...

# Run tests with timeout
test-timeout:
	@echo "Running tests with 5 minute timeout..."
	cd .. && go test -v -timeout=5m ./test/...

# Clean test artifacts
clean:
	@echo "Cleaning test artifacts..."
	cd .. && go clean -testcache
	cd .. && rm -f coverage.out coverage.html
	@echo "Test artifacts cleaned"

# Setup test environment (install dependencies)
setup:
	@echo "Setting up test environment..."
	cd .. && go mod tidy
	cd .. && go mod download
	@echo "Test environment setup complete"

# Lint tests
lint-tests:
	@echo "Linting test files..."
	cd .. && golangci-lint run ./test/...

# Generate test mocks (if using mockgen)
generate-mocks:
	@echo "Generating test mocks..."
	cd .. && go generate ./test/...

# Run tests in Docker (for CI/CD)
test-docker:
	@echo "Running tests in Docker..."
	docker run --rm -v $(PWD)/..:/app -w /app golang:1.24 go test -v ./test/unit/...

# Performance test
test-performance:
	@echo "Running performance tests..."
	cd .. && go test -v -run=TestPerformance ./test/integration/...

# Memory test
test-memory:
	@echo "Running memory tests..."
	cd .. && go test -v -memprofile=mem.prof ./test/unit/...
	cd .. && go tool pprof mem.prof

# CPU profiling test
test-cpu:
	@echo "Running CPU profiling tests..."
	cd .. && go test -v -cpuprofile=cpu.prof ./test/unit/...
	cd .. && go tool pprof cpu.prof

# Continuous testing (watch for changes)
test-watch:
	@echo "Starting continuous testing (requires 'entr' tool)..."
	@echo "Install with: brew install entr (macOS) or apt-get install entr (Ubuntu)"
	find .. -name "*.go" | entr -c make test-unit 