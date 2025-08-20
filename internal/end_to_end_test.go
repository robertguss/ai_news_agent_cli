package internal

import (
        "context"
        "database/sql"
        "net/http"
        "net/http/httptest"
        "os"
        "path/filepath"
        "testing"

        "github.com/robertguss/ai-news-agent-cli/internal/config"
        "github.com/robertguss/ai-news-agent-cli/internal/database"
        "github.com/robertguss/ai-news-agent-cli/internal/fetcher"
        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"
        _ "modernc.org/sqlite"
)

func TestEndToEndConfigFetchStore(t *testing.T) {
        arsRSSContent := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
    <channel>
        <title>Biz &amp; IT – Ars Technica</title>
        <description>Serving the Technologist since 1998. News, reviews, and analysis.</description>
        <link>https://arstechnica.com</link>
        <item>
            <title>Is AI really trying to escape human control and blackmail people?</title>
            <link>https://arstechnica.com/information-technology/2025/08/is-ai-really-trying-to-escape-human-control-and-blackmail-people/</link>
            <pubDate>Wed, 13 Aug 2025 20:28:20 +0000</pubDate>
            <description>Recent AI safety research reveals concerning behaviors in advanced models.</description>
        </item>
    </channel>
</rss>`

        openaiRSSContent := `<?xml version="1.0" encoding="UTF-8"?>
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

        arsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Content-Type", "application/rss+xml")
                w.WriteHeader(http.StatusOK)
                _, _ = w.Write([]byte(arsRSSContent))
        }))
        defer arsServer.Close()

        openaiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Content-Type", "application/rss+xml")
                w.WriteHeader(http.StatusOK)
                _, _ = w.Write([]byte(openaiRSSContent))
        }))
        defer openaiServer.Close()

        tempDir := t.TempDir()
        configPath := filepath.Join(tempDir, "config.yaml")

        configContent := `sources:
  - name: "Ars Technica AI"
    url: "` + arsServer.URL + `"
    type: "rss"
    priority: 2
  - name: "OpenAI Blog"
    url: "` + openaiServer.URL + `"
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
        require.Len(t, cfg.Sources, 2)

        db, err := sql.Open("sqlite", ":memory:")
        require.NoError(t, err)
        defer db.Close()

        err = createTestSchema(db)
        require.NoError(t, err)

        queries := database.New(db)
        ctx := context.Background()

        totalStored := 0
        for _, source := range cfg.Sources {
                stored, err := fetcher.FetchAndStore(ctx, queries, source, cfg, fetcher.FetchOptions{})
                require.NoError(t, err)
                totalStored += stored

                t.Logf("Stored %d articles from %s", stored, source.Name)
        }

        assert.Equal(t, 3, totalStored)

        dbArticles, err := queries.ListArticles(ctx)
        require.NoError(t, err)
        assert.Len(t, dbArticles, 3)

        sourceNames := make(map[string]int)
        for _, article := range dbArticles {
                sourceNames[article.SourceName.String]++
                assert.Equal(t, "unread", article.Status.String)
                assert.True(t, article.Title.Valid)
                assert.True(t, article.Url.Valid)
                assert.True(t, article.PublishedDate.Valid)
        }

        assert.Equal(t, 1, sourceNames["Ars Technica AI"])
        assert.Equal(t, 2, sourceNames["OpenAI Blog"])

        secondRun := 0
        for _, source := range cfg.Sources {
                stored, err := fetcher.FetchAndStore(ctx, queries, source, cfg, fetcher.FetchOptions{})
                require.NoError(t, err)
                secondRun += stored
        }

        assert.Equal(t, 0, secondRun, "Second run should not store duplicates")

        finalArticles, err := queries.ListArticles(ctx)
        require.NoError(t, err)
        assert.Len(t, finalArticles, 3, "Should still have only 3 articles after second run")
}

func createTestSchema(db *sql.DB) error {
        schema := `CREATE TABLE IF NOT EXISTS articles (
                id INTEGER PRIMARY KEY,
                title TEXT,
                url TEXT UNIQUE,
                source_name TEXT,
                published_date DATETIME,
                summary TEXT,
                entities JSON,
                content_type TEXT,
                topics JSON,
                status TEXT DEFAULT 'unread',
                story_group_id TEXT
        );`

        _, err := db.Exec(schema)
        return err
}
