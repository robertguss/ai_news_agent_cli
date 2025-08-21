package scraper

import (
	"context"

	"github.com/robertguss/rss-agent-cli/internal/config"
)

type MockScraper struct {
	content string
	err     error
}

func NewMockScraper(content string, err error) *MockScraper {
	return &MockScraper{
		content: content,
		err:     err,
	}
}

func (m *MockScraper) Scrape(url string) (string, error) {
	return m.content, m.err
}

func (m *MockScraper) ScrapeWithRetry(ctx context.Context, url string, cfg *config.Config) (string, error) {
	return m.content, m.err
}
