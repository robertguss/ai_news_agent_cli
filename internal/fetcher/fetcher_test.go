package fetcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		w.Write([]byte(rssContent))
	}))
	defer server.Close()

	source := Source{
		Name:     "Test Source",
		URL:      server.URL,
		Type:     "rss",
		Priority: 1,
	}

	ctx := context.Background()
	articles, err := Fetch(ctx, source)
	
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

	articles, err := Fetch(ctx, source)
	
	assert.Error(t, err)
	assert.Nil(t, articles)
}

func TestFetch_InvalidRSS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>Not RSS</body></html>"))
	}))
	defer server.Close()

	source := Source{
		Name:     "Test Source",
		URL:      server.URL,
		Type:     "rss",
		Priority: 1,
	}

	ctx := context.Background()
	articles, err := Fetch(ctx, source)
	
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
	articles, err := Fetch(ctx, source)
	
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

	articles, err := Fetch(ctx, source)
	
	assert.Error(t, err)
	assert.Nil(t, articles)
}
