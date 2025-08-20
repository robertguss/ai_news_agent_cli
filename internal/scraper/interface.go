package scraper

import (
        "context"

        "github.com/robertguss/ai-news-agent-cli/internal/config"
)

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
