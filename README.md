# RSS Agent CLI

An AI-powered RSS aggregation CLI that intelligently curates and summarizes content from any RSS feeds. Perfect for terminal enthusiasts who want smart content curation without leaving their command line. While the default configuration focuses on AI/tech news, you can easily configure it for any topics - sports, business, science, or any RSS-enabled content.

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
- ✅ **Universal RSS Support**: Works with any RSS feeds - news, blogs, podcasts, or custom sources
- ✅ **Smart Fetching**: Automatically fetch content from configured RSS sources with per-source article limiting
- ✅ **AI-Powered Processing**: Summarize articles using Google Gemini API for intelligent curation
- ✅ **Local Storage**: SQLite database for offline access and article management
- ✅ **Terminal-Native Reading**: Beautiful markdown rendering for article content
- ✅ **Article Management**: View, read, and open articles with intuitive commands
- ✅ **Flexible Configuration**: YAML-based configuration with environment variable support
- ✅ **Robust Error Handling**: Comprehensive retry logic and error management
- ✅ **Testing**: Comprehensive test suite with mocks and integration tests

### Planned
- 🔄 **Intelligent Deduplication**: Group duplicate stories from different sources
- 🔄 **Priority-Based Sorting**: Enhanced tier-based source prioritization
- 🔄 **Advanced Filtering**: Filter by source, topic, content type, and read status
- 🔄 **OpenAI Integration**: Alternative AI processing option

## Quick Start

### Prerequisites
- Go 1.25 or higher
- Google Gemini API key (get one at [Google AI Studio](https://aistudio.google.com/))
- SQLite (included with most systems)

### Installation

```bash
# Clone and build from source
git clone https://github.com/robertguss/rss-agent-cli.git
cd rss-agent-cli
just build
# or alternatively: go build -o bin/rss-agent-cli .
```

### Setup

1. **Get your Gemini API key** from [Google AI Studio](https://aistudio.google.com/)

2. **Set your API key**:
```bash
export GEMINI_API_KEY="your-gemini-api-key-here"
```

3. **Configure sources** (optional - default config included):
```bash
# Copy and customize the config file
cp configs/config.yaml ~/.rss-agent/config.yaml
```

### Basic Usage

```bash
# Display help
./bin/ai-news-agent-cli --help

# Check version
./bin/ai-news-agent-cli --version

# Fetch latest articles (default: 5 per source)
./bin/ai-news-agent-cli fetch

# Fetch more articles per source
./bin/ai-news-agent-cli fetch -n 10

# Fetch unlimited articles (legacy behavior)
./bin/ai-news-agent-cli fetch -n 0

# View articles
./bin/ai-news-agent-cli view

# Read a specific article
./bin/ai-news-agent-cli read 1

# Open article in browser
./bin/ai-news-agent-cli open 1
```

## Usage

### Available Commands

The CLI supports the following commands:

```bash
# Fetch and process new articles from configured sources
./bin/ai-news-agent-cli fetch

# Fetch with custom article limit per source
./bin/ai-news-agent-cli fetch -n 10        # Max 10 articles per source
./bin/ai-news-agent-cli fetch --limit 3    # Max 3 articles per source

# View stored articles with AI-generated summaries
./bin/ai-news-agent-cli view

# Read full article content in terminal with markdown rendering
./bin/ai-news-agent-cli read <article-number>

# Open article in your default browser
./bin/ai-news-agent-cli open <article-number>

# Generate shell completion scripts
./bin/ai-news-agent-cli completion [bash|zsh|fish|powershell]
```

### Command Options

```bash
# Fetch command options
./bin/ai-news-agent-cli fetch -n 5           # Limit to 5 articles per source (default)
./bin/ai-news-agent-cli fetch --limit 10     # Limit to 10 articles per source
./bin/ai-news-agent-cli fetch -n 0           # Unlimited articles (legacy behavior)

# Read command options
./bin/ai-news-agent-cli read 1 --no-cache    # Force fresh fetch
./bin/ai-news-agent-cli read 1 --no-style    # Plain text output

# View command options (coming soon)
./bin/ai-news-agent-cli view --all           # Show read and unread
./bin/ai-news-agent-cli view --unread        # Show only unread
```

### Example Workflow

1. **Fetch latest articles**: `./bin/ai-news-agent-cli fetch` (gets 5 newest per source by default)
2. **Review AI summaries**: `./bin/ai-news-agent-cli view`
3. **Read interesting articles**: `./bin/ai-news-agent-cli read 2`
4. **Open complex content in browser**: `./bin/ai-news-agent-cli open 5`

### Article Limiting

By default, the fetch command retrieves the **5 most recent articles** from each RSS source to focus on current news. This provides faster fetching and more relevant content compared to processing entire RSS feeds.

- **Default behavior**: `./bin/ai-news-agent-cli fetch` (5 articles per source)
- **Custom limit**: `./bin/ai-news-agent-cli fetch -n 10` (10 articles per source)  
- **Legacy unlimited**: `./bin/ai-news-agent-cli fetch -n 0` (all articles, slower)

Articles are sorted by publish date (newest first) before applying the limit.

## Configuration

The application uses a combination of configuration files and environment variables:

### Environment Variables

```bash
# Required: Gemini API key
export GEMINI_API_KEY="your-gemini-api-key-here"

# Optional: Override default timeouts and retry settings
export NETWORK_TIMEOUT="30s"
export MAX_RETRIES="3"
export BACKOFF_BASE_MS="250"
export BACKOFF_MAX_MS="2000"
export DB_BUSY_RETRIES="3"

# Optional: Custom log file location
export LOG_FILE="$HOME/.rss-agent/agent.log"
```

### Configuration File

The application looks for `config.yaml` in the current directory or `./configs/` directory:

```yaml
# Database file location
dsn: "./rss-agent.db"

# RSS sources to fetch from - customize for any content type!
sources:
  # AI/Tech News (default configuration)
  - name: "Hugging Face Blog"
    url: "https://huggingface.co/blog/feed.xml"
    type: "rss"
    priority: 1
  - name: "OpenAI Blog"
    url: "https://openai.com/blog/rss.xml"
    type: "rss"
    priority: 1
  - name: "Ars Technica AI"
    url: "http://feeds.arstechnica.com/arstechnica/technology-lab"
    type: "rss"
    priority: 2

  # Example: Sports feeds
  # - name: "ESPN NFL"
  #   url: "https://www.espn.com/espn/rss/nfl/news"
  #   type: "rss"
  #   priority: 1

  # Example: Business news
  # - name: "Reuters Business"
  #   url: "https://feeds.reuters.com/reuters/businessNews"
  #   type: "rss"
  #   priority: 1

  # Example: Programming blogs
  # - name: "Hacker News"
  #   url: "https://hnrss.org/frontpage"
  #   type: "rss"
  #   priority: 2

# Optional: Override default settings
network_timeout: "30s"
max_retries: 3
backoff_base_ms: 250
backoff_max_ms: 2000
db_busy_retries: 3
log_file: "$HOME/.rss-agent/agent.log"
```

### Source Priority System

- **Priority 1**: High-priority sources (official blogs, research institutions)
- **Priority 2**: Medium-priority sources (tech news sites)
- **Priority 3**: Lower-priority sources (general tech media)

## Project Structure

```
rss-agent-cli/
├── cmd/                           # CLI commands (Cobra)
│   ├── fetch.go                  # Fetch articles command
│   ├── open.go                   # Open article in browser
│   ├── read.go                   # Read article in terminal
│   ├── root.go                   # Root command and version
│   ├── view.go                   # View articles list
│   └── *_test.go                 # Command tests
├── internal/                      # Internal packages
│   ├── ai/                       # AI processing
│   │   └── processor/            # AI processor implementations
│   ├── article/                  # Article operations
│   ├── browserutil/              # Browser utilities
│   ├── config/                   # Configuration management
│   ├── database/                 # SQLite operations and schema
│   ├── fetcher/                  # RSS content fetching
│   ├── health/                   # Health check utilities
│   ├── scraper/                  # Web content scraping
│   ├── state/                    # Application state management
│   ├── testutil/                 # Testing utilities
│   └── tui/                      # Terminal UI components
├── pkg/                          # Public packages
│   ├── errs/                     # Error handling utilities
│   ├── logging/                  # Logging utilities
│   └── retry/                    # Retry logic utilities
├── configs/                      # Configuration files
│   └── config.yaml              # Default configuration
├── docs/                         # Documentation and specifications
├── main.go                      # Application entry point
├── go.mod                       # Go module definition
└── README.md                    # This file
```

### RSS Feed Examples

The RSS Agent CLI works with any RSS feeds. Here are some popular categories:

**Technology & Programming:**
```yaml
sources:
  - name: "Hacker News"
    url: "https://hnrss.org/frontpage"
    type: "rss"
    priority: 1
  - name: "GitHub Blog"
    url: "https://github.blog/feed/"
    type: "rss"
    priority: 1
  - name: "Stack Overflow Blog"
    url: "https://stackoverflow.blog/feed/"
    type: "rss"
    priority: 2
```

**Business & Finance:**
```yaml
sources:
  - name: "Reuters Business"
    url: "https://feeds.reuters.com/reuters/businessNews"
    type: "rss"
    priority: 1
  - name: "TechCrunch"
    url: "https://techcrunch.com/feed/"
    type: "rss"
    priority: 1
```

**Science & Research:**
```yaml
sources:
  - name: "Nature News"
    url: "https://www.nature.com/nature.rss"
    type: "rss"
    priority: 1
  - name: "Science Daily"
    url: "https://www.sciencedaily.com/rss/top.xml"
    type: "rss"
    priority: 1
```

**Sports:**
```yaml
sources:
  - name: "ESPN Top Stories"
    url: "https://www.espn.com/espn/rss/news"
    type: "rss"
    priority: 1
  - name: "BBC Sport"
    url: "http://feeds.bbci.co.uk/sport/rss.xml"
    type: "rss"
    priority: 1
```

## Roadmap

### ✅ Phase 1: Core Infrastructure (Complete)
- ✅ SQLite database schema and operations
- ✅ Configuration management system with YAML and environment variables
- ✅ RSS source definitions and management
- ✅ Comprehensive error handling and retry logic
- ✅ Logging system with configurable output

### ✅ Phase 2: Content Processing (Complete)
- ✅ RSS feed scraping and parsing
- ✅ Web content extraction and cleaning
- ✅ Google Gemini API integration for AI summarization
- ✅ Article storage and retrieval system
- ✅ Content caching and offline access

### ✅ Phase 3: User Interface (Complete)
- ✅ CLI commands for all core operations (fetch, view, read, open)
- ✅ Terminal-based article reading with markdown rendering
- ✅ Browser integration for complex content
- ✅ Shell completion support
- ✅ Interactive TUI for article viewing with filtering and search
- ✅ Source grouping in terminal UI for better organization

### 🔄 Phase 4: Intelligence Features (In Progress)
- ✅ Basic source priority system
- ✅ Story grouping infrastructure (database schema)
- 🔄 Story deduplication and clustering implementation
- 🔄 Advanced topic categorization and filtering
- ✅ Read/unread status tracking
- 🔄 Content similarity detection

### 🔄 Phase 5: Enhanced User Experience (Planned)
- 🔄 Advanced filtering and search capabilities
- 🔄 Export functionality (JSON, markdown)
- 🔄 Automation support (cron-friendly scheduling)
- 🔄 Custom source management via CLI
- 🔄 OpenAI integration as alternative AI provider

### Future Considerations
- [ ] Integration with Discord/Slack for automated digests
- [ ] Machine learning for personalized content ranking
- [ ] Multi-language support
- [ ] Web interface for remote access

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
git clone https://github.com/robertguss/rss-agent-cli.git
cd rss-agent-cli

# Install dependencies
go mod download

# Set up your Gemini API key
export GEMINI_API_KEY="your-gemini-api-key-here"

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Build the application
just build
# or: go build -o bin/rss-agent-cli .

# Run locally
./bin/ai-news-agent-cli --help

# Try the full workflow
./bin/ai-news-agent-cli fetch
./bin/ai-news-agent-cli view
./bin/ai-news-agent-cli read 1
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

**Note**: This project is production-ready for core AI news aggregation workflows. The CLI provides comprehensive functionality for fetching, processing, and reading AI news articles with intelligent summarization. Check the [roadmap](#roadmap) for upcoming enhancements and advanced features.
