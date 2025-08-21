package article

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchArticle(t *testing.T) {
	tests := []struct {
		name            string
		articleURL      string
		noCache         bool
		serverResponse  func(w http.ResponseWriter, r *http.Request)
		expectedError   string
		expectedContent string
	}{
		{
			name:       "successful fetch",
			articleURL: "https://example.com/article",
			noCache:    false,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				assert.Contains(t, r.URL.Path, "https://example.com/article")
				assert.NotContains(t, r.URL.RawQuery, "no-cache=true")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("# Test Article\n\nThis is test content."))
			},
			expectedContent: "# Test Article\n\nThis is test content.",
		},
		{
			name:       "successful fetch with no-cache",
			articleURL: "https://example.com/fresh",
			noCache:    true,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				assert.Contains(t, r.URL.Path, "https://example.com/fresh")
				assert.Contains(t, r.URL.RawQuery, "no-cache=true")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("Fresh content"))
			},
			expectedContent: "Fresh content",
		},
		{
			name:       "HTTP error",
			articleURL: "https://example.com/notfound",
			noCache:    false,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("Not Found"))
			},
			expectedError: "jina reader API returned 404 Not Found",
		},
		{
			name:       "empty response",
			articleURL: "https://example.com/empty",
			noCache:    false,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(""))
			},
			expectedError: "received empty content from jina reader",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			originalEndpoint := SetJinaEndpointForTesting(server.URL + "/%s")
			defer SetJinaEndpointForTesting(originalEndpoint)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			content, err := FetchArticle(ctx, tt.articleURL, tt.noCache)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedContent, content)
			}
		})
	}
}

func TestFetchArticleTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Too slow"))
	}))
	defer server.Close()

	originalEndpoint := SetJinaEndpointForTesting(server.URL + "/%s")
	defer SetJinaEndpointForTesting(originalEndpoint)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := FetchArticle(ctx, "https://example.com/slow", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch content")
}

func TestFetchArticleURLEncoding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		t.Logf("Received path: %s", path)
		assert.Contains(t, path, "example.com")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Encoded URL content"))
	}))
	defer server.Close()

	originalEndpoint := SetJinaEndpointForTesting(server.URL + "/%s")
	defer SetJinaEndpointForTesting(originalEndpoint)

	ctx := context.Background()
	content, err := FetchArticle(ctx, "https://example.com/article?id=123&type=news", false)

	require.NoError(t, err)
	assert.Equal(t, "Encoded URL content", content)
}
