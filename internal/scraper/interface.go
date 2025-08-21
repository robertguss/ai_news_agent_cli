package scraper

import (
	"context"

	"github.com/robertguss/rss-agent-cli/internal/config"
)

// Scraper defines the interface for content extraction from web URLs.
type Scraper interface {
	Scrape(url string) (string, error)
	ScrapeWithRetry(ctx context.Context, url string, cfg *config.Config) (string, error)
}

type JinaScraper struct{}

func (j *JinaScraper) Scrape(url string) (string, error) {
	return Scrape(url)
}

func (j *JinaScraper) ScrapeWithRetry(ctx context.Context, url string, cfg *config.Config) (string, error) {
	return ScrapeWithRetry(ctx, url, cfg)
}

func NewJinaScraper() *JinaScraper {
	return &JinaScraper{}
}
