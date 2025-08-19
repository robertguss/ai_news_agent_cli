# AI News Agent CLI

An AI-powered news aggregation CLI that helps software engineers stay up-to-date with the latest developments in Artificial Intelligence. Built for terminal enthusiasts who want intelligent news curation without leaving their command line.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Usage](#usage)
- [Configuration](#configuration)
- [Project Structure](#project-structure)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgements](#acknowledgements)

## Features

### Current
- ✅ Basic CLI scaffold with Cobra framework
- ✅ Version information and help system

### Planned
- 🔄 **Smart Fetching**: Automatically fetch content from curated AI news sources
- 🔄 **AI-Powered Processing**: Summarize articles using OpenAI/Gemini APIs
- 🔄 **Intelligent Deduplication**: Group duplicate stories from different sources
- 🔄 **Priority-Based Sorting**: Tier-based source prioritization (Official blogs > Research > General tech media)
- 🔄 **Terminal-Native Reading**: Beautiful markdown rendering with Glow
- 🔄 **Local Storage**: SQLite database for offline access and read tracking
- 🔄 **Flexible Filtering**: Filter by source, topic, content type, and read status

## Quick Start

### Prerequisites
- Go 1.25 or higher
- OpenAI API key or Google Gemini API key
- SQLite (for local storage)

### Installation

```bash
go install github.com/robertguss/ai-news-agent-cli@latest
```

### Basic Usage

```bash
# Display help
ai-news-agent-cli --help

# Check version
ai-news-agent-cli --version

# Current functionality
ai-news-agent-cli
```

## Usage

### Planned Commands

Once fully implemented, the CLI will support these workflows:

```bash
# Fetch and process new AI news
ai-news-agent-cli fetch

# View unread news summaries
ai-news-agent-cli view

# View all news (read and unread)
ai-news-agent-cli view --all

# Filter by specific criteria
ai-news-agent-cli view --source "Google AI Blog"
ai-news-agent-cli view --topic "Large Language Models"
ai-news-agent-cli view --type "Research Paper"

# Read full article in terminal
ai-news-agent-cli read 3

# Open article in browser
ai-news-agent-cli open 3
```

### Example Workflow

1. **Fetch latest news**: `ai-news-agent-cli fetch`
2. **Review summaries**: `ai-news-agent-cli view`
3. **Read interesting articles**: `ai-news-agent-cli read 2`
4. **Open complex content in browser**: `ai-news-agent-cli open 5`

## Configuration

The application will use environment variables for configuration:

```bash
# Required: AI API credentials (choose one)
export OPENAI_API_KEY="your-openai-api-key"
export GEMINI_API_KEY="your-gemini-api-key"

# Optional: Database location
export AI_NEWS_DB_PATH="$HOME/.ai-news-agent/news.db"

# Optional: External tool configurations
export JINA_READER_URL="https://r.jina.ai"
export GLOW_STYLE="dark"
```

## Project Structure

```
ai-news-agent-cli/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go            # Root command and version
│   └── root_test.go       # Command tests
├── internal/              # Internal packages
│   ├── database/          # SQLite operations (planned)
│   ├── fetcher/           # Content fetching (planned)
│   ├── processor/         # AI processing (planned)
│   ├── scraper/           # Source scraping (planned)
│   └── config/            # Configuration management (planned)
├── docs/                  # Documentation and specifications
├── main.go               # Application entry point
├── go.mod                # Go module definition
└── README.md             # This file
```

## Roadmap

### Phase 1: Core Infrastructure
- [ ] SQLite database schema and operations
- [ ] Configuration management system
- [ ] Basic source definitions and management

### Phase 2: Content Processing
- [ ] Web scraping for various source types
- [ ] Integration with Jina Reader for clean content extraction
- [ ] OpenAI/Gemini API integration for summarization
- [ ] Entity extraction (organizations, products, people)

### Phase 3: Intelligence Features
- [ ] Story deduplication and clustering
- [ ] Source priority system implementation
- [ ] Topic categorization and filtering
- [ ] Read/unread status tracking

### Phase 4: User Experience
- [ ] Terminal UI improvements with Glow integration
- [ ] Advanced filtering and search capabilities
- [ ] Export functionality (JSON, markdown)
- [ ] Automation support (cron-friendly)

### Future Considerations
- [ ] Integration with Discord/Slack for automated digests
- [ ] Custom source management via CLI
- [ ] Machine learning for personalized content ranking
- [ ] Multi-language support

## Contributing

We welcome contributions! Here's how to get started:

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/your-feature-name`
3. **Make your changes** following the project conventions
4. **Run tests**: `go test ./...`
5. **Run linting**: `go fmt ./...` and `go vet ./...`
6. **Commit your changes**: `git commit -m "feat: add your feature"`
7. **Push to your fork**: `git push origin feature/your-feature-name`
8. **Create a Pull Request**

### Development Setup

```bash
# Clone the repository
git clone https://github.com/robertguss/ai-news-agent-cli.git
cd ai-news-agent-cli

# Install dependencies
go mod download

# Run tests
go test ./...

# Build the application
go build -o ai-news-agent-cli main.go

# Run locally
./ai-news-agent-cli
```

### Code Style

- Follow standard Go conventions
- Use `go fmt` for formatting
- Run `go vet` for static analysis
- Write tests for new functionality
- Keep functions focused and well-documented

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgements

This project builds upon excellent open-source tools and services:

- **[Cobra](https://github.com/spf13/cobra)** - Powerful CLI framework for Go
- **[OpenAI](https://openai.com/)** - AI-powered content processing
- **[Google Gemini](https://ai.google.dev/)** - Alternative AI processing option
- **[Jina AI](https://jina.ai/)** - Clean content extraction via Jina Reader
- **[Glow](https://github.com/charmbracelet/glow)** - Beautiful terminal markdown rendering
- **[SQLite](https://sqlite.org/)** - Reliable local data storage

---

**Note**: This project is currently in early development. The CLI currently provides basic functionality while the full feature set is being implemented. Check the [roadmap](#roadmap) for planned features and development progress.
