package testutil

import (
	"time"

	"github.com/robertguss/rss-agent-cli/internal/config"
)

func TestConfig() *config.Config {
	return &config.Config{
		DSN:            ":memory:",
		NetworkTimeout: 10 * time.Second,
		MaxRetries:     3,
		BackoffBaseMs:  250,
		BackoffMaxMs:   2000,
		DBBusyRetries:  3,
		Sources: []config.Source{
			{
				Name:     "test-source",
				URL:      "https://example.com/feed.xml",
				Type:     "rss",
				Priority: 1,
			},
		},
	}
}
