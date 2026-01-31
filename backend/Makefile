.PHONY: build test test-verbose test-cover clean run setup-test-db help

# Build the application
build:
	go build -o lifesupport-api

# Run the application
run: build
	./lifesupport-api

# Run all tests
test:
	go test ./...

# Run tests with verbose output
test-verbose:
	go test ./... -v

# Run API tests only
test-api:
	go test ./pkg/api -v

# Run storer tests only
test-storer:
	go test ./pkg/storer -v

# Run tests with coverage
test-cover:
	go test ./... -cover

# Generate coverage report
coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Setup test database
setup-test-db:
	./setup-test-db.sh

# Clean build artifacts
clean:
	rm -f lifesupport-api
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	go vet ./...

# Run all checks before commit
check: fmt lint test

# Display help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  run            - Build and run the application"
	@echo "  test           - Run all tests"
	@echo "  test-verbose   - Run tests with verbose output"
	@echo "  test-api       - Run API package tests only"
	@echo "  test-storer    - Run storer package tests only"
	@echo "  test-cover     - Run tests with coverage"
	@echo "  coverage       - Generate HTML coverage report"
	@echo "  setup-test-db  - Create test database"
	@echo "  clean          - Remove build artifacts"
	@echo "  deps           - Download and tidy dependencies"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter"
	@echo "  check          - Run fmt, lint, and test"
	@echo "  help           - Display this help message"
