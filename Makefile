# AI News Agent CLI Makefile

.PHONY: build test lint clean fmt vet coverage help

# Default target
all: fmt vet test build

# Build the application
build:
	@echo "Building rss-agent-cli..."
	go build -o rss-agent-cli main.go

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	go test -cover ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run static analysis
vet:
	@echo "Running go vet..."
	go vet ./...

# Run linting (requires golangci-lint)
lint:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f rss-agent-cli

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run the application
run: build
	@echo "Running rss-agent-cli..."
	./rss-agent-cli

# Show help
help:
	@echo "Available targets:"
	@echo "  build     - Build the application"
	@echo "  test      - Run tests"
	@echo "  coverage  - Run tests with coverage"
	@echo "  fmt       - Format code"
	@echo "  vet       - Run static analysis"
	@echo "  lint      - Run linting (requires golangci-lint)"
	@echo "  clean     - Clean build artifacts"
	@echo "  deps      - Install dependencies"
	@echo "  run       - Build and run the application"
	@echo "  help      - Show this help message"
