package internal

import (
        "context"
        "net/http"
        "net/http/httptest"
        "os"
        "path/filepath"
        "testing"

        "github.com/robertguss/ai-news-agent-cli/internal/config"
        "github.com/robertguss/ai-news-agent-cli/internal/fetcher"
        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"
)

func TestConfigAndFetcherIntegration(t *testing.T) {
        rssContent := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
    <channel>
        <title>OpenAI News</title>
        <description>The OpenAI blog</description>
        <link>https://openai.com/news</link>
        <item>
            <title>Introducing GPT-5 for developers</title>
            <description>Introducing GPT-5 in our API platform—offering high reasoning performance, new controls for devs, and best-in-class results on real coding tasks.</description>
            <link>https://openai.com/index/introducing-gpt-5-for-developers</link>
            <pubDate>Thu, 07 Aug 2025 10:00:00 GMT</pubDate>
        </item>
        <item>
            <title>GPT-5 and the new era of work</title>
            <description>GPT-5 is OpenAI's most advanced model—transforming enterprise AI, automation, and workforce productivity in the new era of intelligent work.</description>
            <link>https://openai.com/index/gpt-5-new-era-of-work</link>
            <pubDate>Thu, 07 Aug 2025 10:00:00 GMT</pubDate>
        </item>
    </channel>
</rss>`

        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Content-Type", "application/rss+xml")
                w.WriteHeader(http.StatusOK)
                _, _ = w.Write([]byte(rssContent))
        }))
        defer server.Close()

        tempDir := t.TempDir()
        configPath := filepath.Join(tempDir, "config.yaml")

        configContent := `sources:
  - name: "Test OpenAI Blog"
    url: "` + server.URL + `"
    type: "rss"
    priority: 1`

        err := os.WriteFile(configPath, []byte(configContent), 0644)
        require.NoError(t, err)

        originalDir, err := os.Getwd()
        require.NoError(t, err)
        defer func() { _ = os.Chdir(originalDir) }()

        err = os.Chdir(tempDir)
        require.NoError(t, err)

        cfg, err := config.Load()
        require.NoError(t, err)
        require.Len(t, cfg.Sources, 1)

        source := cfg.Sources[0]
        assert.Equal(t, "Test OpenAI Blog", source.Name)
        assert.Equal(t, server.URL, source.URL)
        assert.Equal(t, "rss", source.Type)
        assert.Equal(t, 1, source.Priority)

        ctx := context.Background()
        articles, err := fetcher.Fetch(ctx, source, cfg, fetcher.FetchOptions{})
        require.NoError(t, err)
        require.Len(t, articles, 2)

        assert.Equal(t, "Introducing GPT-5 for developers", articles[0].Title)
        assert.Equal(t, "https://openai.com/index/introducing-gpt-5-for-developers", articles[0].Link)
        assert.False(t, articles[0].PublishedDate.IsZero())

        assert.Equal(t, "GPT-5 and the new era of work", articles[1].Title)
        assert.Equal(t, "https://openai.com/index/gpt-5-new-era-of-work", articles[1].Link)
        assert.False(t, articles[1].PublishedDate.IsZero())
}
