package fetcher

import (
        "context"
        "database/sql"
        "net/http"
        "net/http/httptest"
        "testing"
        "time"

        "github.com/robertguss/ai-news-agent-cli/internal/database"
        "github.com/robertguss/ai-news-agent-cli/internal/testutil"
        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"
        _ "modernc.org/sqlite"
)

func TestStoreArticles_Success(t *testing.T) {
        db, err := sql.Open("sqlite", ":memory:")
        require.NoError(t, err)
        defer db.Close()

        err = createTestSchema(db)
        require.NoError(t, err)

        queries := database.New(db)

        articles := []Article{
                {
                        Title:         "Test Article 1",
                        Link:          "https://example.com/article1",
                        PublishedDate: time.Now(),
                },
                {
                        Title:         "Test Article 2",
                        Link:          "https://example.com/article2",
                        PublishedDate: time.Now().Add(-1 * time.Hour),
                },
        }

        source := Source{
                Name:     "Test Source",
                URL:      "https://example.com/feed",
                Type:     "rss",
                Priority: 1,
        }

        cfg := testutil.TestConfig()
        ctx := context.Background()
        stored, err := StoreArticles(ctx, queries, articles, source, cfg)
        require.NoError(t, err)
        assert.Equal(t, 2, stored)

        dbArticles, err := queries.ListArticles(ctx)
        require.NoError(t, err)
        assert.Len(t, dbArticles, 2)

        assert.Equal(t, "Test Article 1", dbArticles[0].Title.String)
        assert.Equal(t, "https://example.com/article1", dbArticles[0].Url.String)
        assert.Equal(t, "Test Source", dbArticles[0].SourceName.String)
        assert.Equal(t, "unread", dbArticles[0].Status.String)
}

func TestStoreArticles_DuplicateHandling(t *testing.T) {
        db, err := sql.Open("sqlite", ":memory:")
        require.NoError(t, err)
        defer db.Close()

        err = createTestSchema(db)
        require.NoError(t, err)

        queries := database.New(db)

        articles := []Article{
                {
                        Title:         "Duplicate Article",
                        Link:          "https://example.com/duplicate",
                        PublishedDate: time.Now(),
                },
        }

        source := Source{
                Name:     "Test Source",
                URL:      "https://example.com/feed",
                Type:     "rss",
                Priority: 1,
        }

        cfg := testutil.TestConfig()
        ctx := context.Background()

        stored1, err := StoreArticles(ctx, queries, articles, source, cfg)
        require.NoError(t, err)
        assert.Equal(t, 1, stored1)

        stored2, err := StoreArticles(ctx, queries, articles, source, cfg)
        require.NoError(t, err)
        assert.Equal(t, 0, stored2)

        dbArticles, err := queries.ListArticles(ctx)
        require.NoError(t, err)
        assert.Len(t, dbArticles, 1)
}

func TestStoreArticles_EmptyList(t *testing.T) {
        db, err := sql.Open("sqlite", ":memory:")
        require.NoError(t, err)
        defer db.Close()

        err = createTestSchema(db)
        require.NoError(t, err)

        queries := database.New(db)

        var articles []Article
        source := Source{
                Name:     "Test Source",
                URL:      "https://example.com/feed",
                Type:     "rss",
                Priority: 1,
        }

        cfg := testutil.TestConfig()
        ctx := context.Background()
        stored, err := StoreArticles(ctx, queries, articles, source, cfg)
        require.NoError(t, err)
        assert.Equal(t, 0, stored)
}

func TestFetchAndStore_Integration(t *testing.T) {
        rssContent := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
    <channel>
        <title>Test Feed</title>
        <item>
            <title>Integration Test Article</title>
            <link>https://example.com/integration</link>
            <pubDate>Wed, 13 Aug 2025 20:28:20 +0000</pubDate>
        </item>
    </channel>
</rss>`

        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Content-Type", "application/rss+xml")
                w.WriteHeader(http.StatusOK)
                _, _ = w.Write([]byte(rssContent))
        }))
        defer server.Close()

        db, err := sql.Open("sqlite", ":memory:")
        require.NoError(t, err)
        defer db.Close()

        err = createTestSchema(db)
        require.NoError(t, err)

        queries := database.New(db)

        source := Source{
                Name:     "Integration Test Source",
                URL:      server.URL,
                Type:     "rss",
                Priority: 1,
        }

        cfg := testutil.TestConfig()
        ctx := context.Background()
        stored, err := FetchAndStore(ctx, queries, source, cfg, FetchOptions{})
        require.NoError(t, err)
        assert.Equal(t, 1, stored)

        dbArticles, err := queries.ListArticles(ctx)
        require.NoError(t, err)
        assert.Len(t, dbArticles, 1)
        assert.Equal(t, "Integration Test Article", dbArticles[0].Title.String)
        assert.Equal(t, "https://example.com/integration", dbArticles[0].Url.String)
        assert.Equal(t, "Integration Test Source", dbArticles[0].SourceName.String)
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
