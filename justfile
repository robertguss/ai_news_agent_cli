# justfile for ai-news-agent-cli
set dotenv-load

BINARY_NAME := "ai-news-agent-cli"
BIN_DIR := "bin"
DIST_DIR := "dist"
VERSION := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`
LDFLAGS := "-ldflags \"-X main.version=" + VERSION + "\""

default:
    @echo "Available recipes:"
    @echo "  build         Build the binary for current platform"
    @echo "  cross-compile Build binaries for multiple platforms"
    @echo "  test          Run tests"
    @echo "  test-coverage Run tests with coverage"
    @echo "  fmt           Format code"
    @echo "  vet           Run go vet"
    @echo "  lint          Run golangci-lint (if available)"
    @echo "  deps          Download dependencies"
    @echo "  generate-mocks Generate mocks using mockery"
    @echo "  clean         Clean build artifacts"
    @echo "  install       Install binary to GOPATH/bin"
    @echo "  run           Run the built binary"
    @echo "  dev           Run with live reload (air) or fallback to go run"

_ensure-bin-dir:
    mkdir -p {{BIN_DIR}}

_ensure-dist-dir:
    mkdir -p {{DIST_DIR}}

build: _ensure-bin-dir
    go build {{LDFLAGS}} -o {{BIN_DIR}}/{{BINARY_NAME}} .

cross-compile: _ensure-dist-dir
    @echo "Building for multiple platforms..."
    GOOS=linux GOARCH=amd64 go build {{LDFLAGS}} -o {{DIST_DIR}}/{{BINARY_NAME}}-linux-amd64 .
    GOOS=linux GOARCH=arm64 go build {{LDFLAGS}} -o {{DIST_DIR}}/{{BINARY_NAME}}-linux-arm64 .
    GOOS=darwin GOARCH=amd64 go build {{LDFLAGS}} -o {{DIST_DIR}}/{{BINARY_NAME}}-darwin-amd64 .
    GOOS=darwin GOARCH=arm64 go build {{LDFLAGS}} -o {{DIST_DIR}}/{{BINARY_NAME}}-darwin-arm64 .
    GOOS=windows GOARCH=amd64 go build {{LDFLAGS}} -o {{DIST_DIR}}/{{BINARY_NAME}}-windows-amd64.exe .
    GOOS=windows GOARCH=arm64 go build {{LDFLAGS}} -o {{DIST_DIR}}/{{BINARY_NAME}}-windows-arm64.exe .
    @echo "Cross-compilation complete. Binaries available in {{DIST_DIR}}/"

test:
    go test -v ./...

test-coverage:
    go test -cover ./...

fmt:
    go fmt ./...

vet:
    go vet ./...

lint:
    #!/usr/bin/env bash
    if command -v golangci-lint >/dev/null 2>&1; then
        golangci-lint run
    else
        echo "golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    fi

deps:
    go mod download
    go mod tidy

generate-mocks:
    go generate ./...

clean:
    rm -rf {{BIN_DIR}}
    rm -rf {{DIST_DIR}}
    go clean

install: build
    go install {{LDFLAGS}} .

run: build
    ./{{BIN_DIR}}/{{BINARY_NAME}}

dev:
    #!/usr/bin/env bash
    if command -v air >/dev/null 2>&1; then
        air
    else
        echo "Air not found. Install it with: go install github.com/air-verse/air@latest"
        echo "Running without live reload..."
        go run .
    fi
