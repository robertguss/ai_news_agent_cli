# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Commands

### Build and Run
- **Build**: `just build` or `go build -o bin/rss-agent-cli .`
- **Run built binary**: `just run` or `./bin/rss-agent-cli`
- **Dev with live reload**: `just dev` (uses air if available, fallback to go run)
- **Cross-compile**: `just cross-compile` (builds for multiple platforms)
- **Install**: `just install` (installs to GOPATH/bin)

### Testing
- **Run all tests**: `go test ./...`
- **Verbose tests**: `go test -v ./...` 
- **Test single package**: `go test ./internal/database/...` (replace with specific package)
- **Test with coverage**: `go test -cover ./...`
- **Generate mocks**: `just generate-mocks` or `go generate ./...`

### Code Quality
- **Format**: `go fmt ./...` or `just fmt`
- **Vet**: `go vet ./...` or `just vet` 
- **Lint**: `golangci-lint run` or `just lint` (installs if missing)
- **All quality checks**: `just fmt && just vet && just lint`

### Dependencies and Cleanup
- **Install deps**: `go mod download && go mod tidy` or `just deps`
- **Clean**: `just clean` (removes bin/, dist/, *.db files)

## High-Level Architecture

### Core Concepts
This is an AI-powered news aggregation CLI tool that fetches RSS feeds, processes content with Google Gemini AI, and stores articles in SQLite for terminal-based reading.

### Key Components
- **CLI Framework**: Cobra-based commands in `cmd/` (fetch, view, read, open)
- **Database**: SQLite with sqlc-generated type-safe queries in `internal/database/`
- **AI Processing**: Google Generative AI (Gemini) integration in `internal/ai/processor/`
- **Content Fetching**: RSS parsing with `internal/fetcher/` and web scraping via `internal/scraper/`
- **Terminal UI**: Bubble Tea framework components in `internal/tui/`
- **Configuration**: YAML + environment variable support in `internal/config/`

### Data Flow
1. **Fetch**: RSS sources → Parse feeds → Extract articles with metadata
2. **Process**: Web scraping → AI analysis (summary, entities, topics) → Store in SQLite
3. **View**: Query database → Display in terminal UI with summaries
4. **Read**: Fetch full content → Render markdown in terminal

### Database Schema
Single `articles` table with columns: id, title, url (unique), source_name, published_date, summary, entities (JSON), content_type, topics (JSON), status, story_group_id.

### Configuration Requirements
- **Required**: `GEMINI_API_KEY` environment variable
- **Optional**: Config file at `./config.yaml` or `./configs/config.yaml` for RSS sources
- **Defaults**: Network timeouts, retry policies, database location can be overridden via env vars

### Testing Approach
- Uses testify for assertions and mockery for interface mocking
- Integration tests in `*_integration_test.go` files
- Mocks generated in `internal/*/mocks/` directories
- End-to-end tests verify full pipeline functionality

### Error Handling Patterns
- Custom error wrapping with `fmt.Errorf("context: %w", err)`
- Retry logic with exponential backoff in `pkg/retry/`
- Graceful degradation for network and AI processing failures

### Important Implementation Details
- Article limiting: Default 5 articles per RSS source (configurable via `-n` flag)
- Concurrent fetching with goroutines and sync.WaitGroup
- Terminal detection and graceful fallback for non-TTY environments
- Browser integration for complex content that doesn't render well in terminal