# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

AI News Agent CLI is a command-line application that acts as an intelligent AI agent to help users stay up-to-date with AI news. The tool fetches, processes, summarizes, and displays content from various sources using a "hacker-friendly" terminal interface.

## Architecture

This Go-based CLI application follows a modular architecture:

- **CLI Framework**: Cobra for command/subcommand management
- **Configuration**: Viper for YAML config handling  
- **Database**: SQLite with `sqlc` for type-safe Go code generation
- **AI Processing**: Cloud AI API (OpenAI GPT/Google Gemini) for NLP tasks
- **Content Extraction**: Jina Reader for clean markdown from URLs
- **Terminal UI**: Lip Gloss for styled output, Glow for markdown rendering
- **Testing**: Go's built-in testing with testify for assertions, mockery for mocks

### Project Structure
```
ai-news-agent/
├── cmd/           # Cobra commands (root, fetch, view, read, open)
├── internal/      # Internal application logic
│   ├── database/  # SQLite schema and sqlc-generated code
│   ├── fetcher/   # RSS/content fetching logic
│   ├── processor/ # AI processing interfaces and implementations
│   ├── scraper/   # Content extraction from URLs
│   └── config/    # Configuration loading
├── docs/          # Project specifications and plans
├── config.yaml    # Source configurations (when created)
├── main.go        # Application entry point
└── go.mod
```

## Development Commands

### Project Setup
```bash
# Initialize Go module
go mod init github.com/user/ai-news-agent

# Add core dependencies
go get github.com/spf13/cobra@latest
go get github.com/stretchr/testify@latest
go get modernc.org/sqlite@latest
go get github.com/spf13/viper@latest
go get github.com/mmcdole/gofeed@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/sashabaranov/go-openai@latest
go get github.com/pkg/browser@latest

# Install development tools
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/vektra/mockery/v2@latest
```

### Database Operations
```bash
# Generate type-safe database code (run from internal/database/)
sqlc generate

# Generate mocks for interfaces
go generate ./...
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./cmd/...
go test ./internal/database/...

# Run tests with coverage
go test -cover ./...
```

### Building and Running
```bash
# Build the application
go build -o ai-news .

# Run directly
go run main.go

# Format and vet code
go fmt ./...
go vet ./...
```

## Key Commands (Planned)

### Core CLI Commands
- `ai-news fetch` - Fetches and processes new content from configured sources
- `ai-news view` - Displays processed news with filtering options (--all, --source, --topic)
- `ai-news read <number>` - Renders full article content in terminal using Glow
- `ai-news open <number>` - Opens original article in web browser

## Development Phases

The project follows a 5-phase TDD approach:

1. **Phase 1**: Project skeleton, Cobra CLI setup, SQLite database with sqlc
2. **Phase 2**: Configuration loading, RSS fetching, basic `fetch` command
3. **Phase 3**: AI processor interface, mock implementation, real API client
4. **Phase 4**: Enhanced `view` command with styling, `read`/`open` commands
5. **Phase 5**: Error handling, documentation, finalization

## Configuration

### Environment Variables
- `OPENAI_API_KEY` - Required for AI processing functionality

### config.yaml Structure
```yaml
sources:
  - name: "Source Name"
    url: "https://example.com/feed"
    type: "rss"
    priority: 1  # 1=Tier 1 (highest), 2=Tier 2, 3=Tier 3
```

## Database Schema

The SQLite database uses a single `articles` table with columns for:
- Core metadata (title, URL, source, published_date)
- AI-generated analysis (summary, entities, topics, content_type)
- Application state (status, story_group_id for deduplication)

## Testing Strategy

- **Unit Tests**: Test individual functions with mocked dependencies
- **Integration Tests**: Test module interactions with test databases
- **E2E Tests**: Test full CLI command workflows

Use `httptest` for mocking HTTP services and temporary databases for database testing.

## External Dependencies

- **Jina Reader**: `r.jina.ai/{url}` for clean article content
- **Glow**: Terminal markdown renderer (must be installed separately)
- **Cloud AI API**: OpenAI or Google Gemini for content analysis

## Development Notes

- Use `sqlc` annotations in SQL files for type-safe code generation
- Implement robust error handling for network operations and API failures
- Follow Go idioms with proper interface usage and error propagation
- Use `Lip Gloss` for terminal styling to create visually appealing "card" formats
- Store article number-to-URL mappings in temp files for interactive commands
