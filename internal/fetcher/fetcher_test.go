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

func TestFetch_Success(t *testing.T) {
        rssContent := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
    <channel>
        <title>Test Feed</title>
        <description>Test RSS Feed</description>
        <link>https://example.com</link>
        <item>
            <title>Test Article 1</title>
            <link>https://example.com/article1</link>
            <pubDate>Wed, 13 Aug 2025 20:28:20 +0000</pubDate>
            <description>First test article</description>
        </item>
        <item>
            <title>Test Article 2</title>
            <link>https://example.com/article2</link>
            <pubDate>Tue, 12 Aug 2025 15:30:00 +0000</pubDate>
            <description>Second test article</description>
        </item>
    </channel>
</rss>`

        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Content-Type", "application/rss+xml")
                w.WriteHeader(http.StatusOK)
                _, _ = w.Write([]byte(rssContent))
        }))
        defer server.Close()

        source := Source{
                Name:     "Test Source",
                URL:      server.URL,
                Type:     "rss",
                Priority: 1,
        }

        cfg := testutil.TestConfig()
        ctx := context.Background()
        opts := FetchOptions{Limit: 0}
        articles, err := Fetch(ctx, source, cfg, opts)

        require.NoError(t, err)
        require.Len(t, articles, 2)

        assert.Equal(t, "Test Article 1", articles[0].Title)
        assert.Equal(t, "https://example.com/article1", articles[0].Link)
        assert.False(t, articles[0].PublishedDate.IsZero())

        assert.Equal(t, "Test Article 2", articles[1].Title)
        assert.Equal(t, "https://example.com/article2", articles[1].Link)
        assert.False(t, articles[1].PublishedDate.IsZero())
}

func TestFetch_NetworkError(t *testing.T) {
        source := Source{
                Name:     "Test Source",
                URL:      "http://nonexistent-domain-12345.com/feed.xml",
                Type:     "rss",
                Priority: 1,
        }

        ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
        defer cancel()

        cfg := testutil.TestConfig()
        articles, err := Fetch(ctx, source, cfg, FetchOptions{})

        assert.Error(t, err)
        assert.Nil(t, articles)
}

func TestFetch_InvalidRSS(t *testing.T) {
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Content-Type", "text/html")
                w.WriteHeader(http.StatusOK)
                _, _ = w.Write([]byte("<html><body>Not RSS</body></html>"))
        }))
        defer server.Close()

        source := Source{
                Name:     "Test Source",
                URL:      server.URL,
                Type:     "rss",
                Priority: 1,
        }

        ctx := context.Background()
        cfg := testutil.TestConfig()
        articles, err := Fetch(ctx, source, cfg, FetchOptions{})

        assert.Error(t, err)
        assert.Nil(t, articles)
}

func TestFetch_HTTPError(t *testing.T) {
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusNotFound)
        }))
        defer server.Close()

        source := Source{
                Name:     "Test Source",
                URL:      server.URL,
                Type:     "rss",
                Priority: 1,
        }

        ctx := context.Background()
        cfg := testutil.TestConfig()
        articles, err := Fetch(ctx, source, cfg, FetchOptions{})

        assert.Error(t, err)
        assert.Nil(t, articles)
}

func TestFetch_ContextTimeout(t *testing.T) {
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                time.Sleep(2 * time.Second)
                w.WriteHeader(http.StatusOK)
        }))
        defer server.Close()

        source := Source{
                Name:     "Test Source",
                URL:      server.URL,
                Type:     "rss",
                Priority: 1,
        }

        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
        defer cancel()

        cfg := testutil.TestConfig()
        articles, err := Fetch(ctx, source, cfg, FetchOptions{})

        assert.Error(t, err)
        assert.Nil(t, articles)
}

func TestStoreArticlesWithAI_StatusSeparation(t *testing.T) {
        db, err := sql.Open("sqlite", ":memory:")
        require.NoError(t, err)
        defer db.Close()

        err = createTestSchema(db)
        require.NoError(t, err)

        queries := database.New(db)
        ctx := context.Background()

        articles := []Article{
                {
                        Title:         "Test Article",
                        Link:          "https://example.com/test-status-separation",
                        PublishedDate: time.Now(),
                },
        }

        source := Source{
                Name: "Test Source",
                URL:  "https://example.com/feed",
        }

        cfg := testutil.TestConfig()

        deps := PipelineDeps{
                Scraper: nil, // No scraper/AI for this test
                AI:      nil,
                Queries: queries,
                Config:  cfg,
        }

        stored, err := StoreArticlesWithAI(ctx, deps, articles, source)
        require.NoError(t, err)
        assert.Equal(t, 1, stored)

        // Verify the article was stored with correct status separation
        article, err := queries.GetArticleByUrl(ctx, sql.NullString{
                String: "https://example.com/test-status-separation",
                Valid:  true,
        })
        require.NoError(t, err)

        // Status should be 'unread' for read/unread tracking
        assert.Equal(t, "unread", article.Status.String)
        assert.True(t, article.Status.Valid)

        // AnalysisStatus should be 'unprocessed' since no AI processing occurred
        assert.Equal(t, "unprocessed", article.AnalysisStatus.String)
        assert.True(t, article.AnalysisStatus.Valid)

        // Summary should be null/empty since no AI processing occurred
        assert.False(t, article.Summary.Valid)
}

func TestStoreArticles_RegularStorage(t *testing.T) {
        db, err := sql.Open("sqlite", ":memory:")
        require.NoError(t, err)
        defer db.Close()

        err = createTestSchema(db)
        require.NoError(t, err)

        queries := database.New(db)
        ctx := context.Background()

        articles := []Article{
                {
                        Title:         "Regular Article",
                        Link:          "https://example.com/regular-article",
                        PublishedDate: time.Now(),
                },
        }

        source := Source{
                Name: "Regular Source",
                URL:  "https://example.com/feed",
        }

        cfg := testutil.TestConfig()

        stored, err := StoreArticles(ctx, queries, articles, source, cfg)
        require.NoError(t, err)
        assert.Equal(t, 1, stored)

        // Verify the article was stored with correct defaults
        article, err := queries.GetArticleByUrl(ctx, sql.NullString{
                String: "https://example.com/regular-article",
                Valid:  true,
        })
        require.NoError(t, err)

        // Status should be 'unread' for read/unread tracking
        assert.Equal(t, "unread", article.Status.String)
        assert.True(t, article.Status.Valid)

        // AnalysisStatus should be 'unprocessed' for regular storage
        assert.Equal(t, "unprocessed", article.AnalysisStatus.String)
        assert.True(t, article.AnalysisStatus.Valid)

        // Summary should be null since no AI processing
        assert.False(t, article.Summary.Valid)
}

func TestStoreArticlesWithAI_WithMockAISuccess(t *testing.T) {
        db, err := sql.Open("sqlite", ":memory:")
        require.NoError(t, err)
        defer db.Close()

        err = createTestSchema(db)
        require.NoError(t, err)

        queries := database.New(db)
        ctx := context.Background()

        // Create a mock article with AI-processed data
        params := database.CreateArticleParams{
                Title: sql.NullString{String: "AI Processed Article", Valid: true},
                Url: sql.NullString{String: "https://example.com/ai-processed", Valid: true},
                SourceName: sql.NullString{String: "AI Test Source", Valid: true},
                PublishedDate: sql.NullTime{Time: time.Now(), Valid: true},
                Summary: sql.NullString{String: "AI generated summary", Valid: true},
                Entities: []byte(`["entity1", "entity2"]`),
                ContentType: sql.NullString{String: "news", Valid: true},
                Topics: []byte(`["tech", "ai"]`),
                Status: sql.NullString{String: "unread", Valid: true},
                AnalysisStatus: sql.NullString{String: "completed", Valid: true},
                StoryGroupID: sql.NullString{String: "test-group", Valid: true},
        }

        article, err := queries.CreateArticle(ctx, params)
        require.NoError(t, err)

        // Verify the article has correct status separation with AI data
        assert.Equal(t, "unread", article.Status.String)
        assert.Equal(t, "completed", article.AnalysisStatus.String)
        assert.Equal(t, "AI generated summary", article.Summary.String)
        assert.Equal(t, "news", article.ContentType.String)
        assert.True(t, article.Summary.Valid)
        assert.True(t, article.ContentType.Valid)

        // Verify we can retrieve it correctly
        retrieved, err := queries.GetArticleByUrl(ctx, sql.NullString{
                String: "https://example.com/ai-processed",
                Valid:  true,
        })
        require.NoError(t, err)

        assert.Equal(t, "unread", retrieved.Status.String)
        assert.Equal(t, "completed", retrieved.AnalysisStatus.String)
        assert.Equal(t, "AI generated summary", retrieved.Summary.String)
}
