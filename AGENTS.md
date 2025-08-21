# AGENTS.md
# Keep this file ≤ 25 lines. Update whenever architecture, testing, or linting changes.

## Commands
- **Build**: `just build` or `go build -o bin/rss-agent-cli .`
- **Test all**: `go test ./...` | **Test verbose**: `go test -v ./...`
- **Test single package**: `go test ./internal/database/...` (replace with specific package)
- **Coverage**: `go test -cover ./...`
- **Lint/Format**: `go fmt ./...` then `go vet ./...` then `golangci-lint run` (if available)
- **Generate mocks**: `mockery --all --output internal/mocks`

## Architecture
- **Language**: Go 1.25+, CLI with Cobra framework (cmd/)
- **Modules**: internal/{database,fetcher,ai/processor,scraper,config,tui,health}
- **Database**: SQLite with sqlc-generated type-safe code in internal/database/
- **UI**: Bubble Tea TUI framework in internal/tui/
- **AI**: Google Generative AI (Gemini) for content processing
- **Testing**: testify + mockery for mocks

## Code Style
- **Packages**: Use internal/ for private modules, cmd/ for CLI commands
- **Errors**: Wrap with `fmt.Errorf("context: %w", err)`, create custom types when needed
- **Imports**: Standard library first, external packages, then local packages
- **Naming**: Follow Go conventions (camelCase vars, PascalCase exports)
- **DB Models**: Use sql.Null* types for nullable fields (sqlc generated)
