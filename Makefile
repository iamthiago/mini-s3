.PHONY: test build clean all coverage coverage-html lint

# Default target
all: build test

# Build all modules
build:
	@echo "Building root module..."
	go build -v ./...
	@echo "Building storage module..."
	cd storage && go build -v ./...

# Test all modules
test:
	@echo "Testing root module..."
	go test -v ./...
	@echo "Testing storage module..."
	cd storage && go test -v ./...

# Clean build artifacts
clean:
	go clean
	cd storage && go clean
	rm -f coverage.out storage/coverage.out coverage.html storage/coverage.html

# Run tests with coverage and generate reports
coverage:
	@echo "Running tests with coverage for root module..."
	go test -coverprofile=coverage.out -covermode=atomic ./...
	@echo "Running tests with coverage for storage module..."
	cd storage && go test -coverprofile=coverage.out -covermode=atomic ./...

# Generate HTML coverage reports
coverage-html: coverage
	@echo "Generating HTML coverage report for root module..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Generating HTML coverage report for storage module..."
	cd storage && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage reports generated: coverage.html and storage/coverage.html"

# Lint all modules
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install with: brew install golangci-lint or go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	@echo "Linting root module..."
	golangci-lint run ./...
	@echo "Linting storage module..."
	cd storage && golangci-lint run ./...