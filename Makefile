.PHONY: test build clean all coverage coverage-html lint fmt fmt-check

# Default target
all: build test

# Build the project
build:
	@echo "Building mini-s3..."
	go build -v ./...

# Test all packages
test:
	@echo "Testing mini-s3..."
	go test -v ./...

# Clean build artifacts
clean:
	go clean
	rm -f coverage.out coverage.html
	rm -rf ./data/*

# Run tests with coverage and generate reports
coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out -covermode=atomic ./...

# Generate HTML coverage report
coverage-html: coverage
	@echo "Generating HTML coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Lint the project
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install with: brew install golangci-lint or go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	@echo "Linting mini-s3..."
	golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	go fmt ./...

fmt-check:
	@echo "Checking code formatting..."
	@test -z "$$(gofmt -s -l .)" || (echo "The following files are not formatted:\n$$(gofmt -s -l .)\nRun 'make fmt' to fix" && exit 1)
