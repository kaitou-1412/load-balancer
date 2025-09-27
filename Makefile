# Go project Makefile

# Variables
BINARY_NAME=load-balancer
GO_FILES=$(shell find . -name "*.go" -type f)

# Default target
.DEFAULT_GOAL := help

# Run the application
run:
	go run .

# Build the application
build:
	go build -o $(BINARY_NAME) .

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint to be installed)
lint:
	golangci-lint run

# Tidy up dependencies
tidy:
	go mod tidy

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	go mod download

# Display help
help:
	@echo "Available commands:"
	@echo "  run           - Run the application"
	@echo "  build         - Build the application"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code (requires golangci-lint)"
	@echo "  tidy          - Tidy up dependencies"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download dependencies"
	@echo "  help          - Show this help message"

.PHONY: run build test test-coverage fmt lint tidy clean deps help
