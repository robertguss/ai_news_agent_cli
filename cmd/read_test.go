package cmd

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/robertguss/ai-news-agent-cli/internal/article"
	"github.com/robertguss/ai-news-agent-cli/internal/state"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		setupState     func() *state.ViewState
		expectedError  string
		expectedOutput string
	}{
		{
			name:          "invalid article number",
			args:          []string{"abc"},
			setupState:    func() *state.ViewState { return &state.ViewState{Articles: map[string]state.ArticleRef{}} },
			expectedError: `invalid article number "abc": must be a positive integer`,
		},
		{
			name:          "no state file",
			args:          []string{"1"},
			setupState:    func() *state.ViewState { return &state.ViewState{Articles: map[string]state.ArticleRef{}} },
			expectedError: "no viewed articles found - run 'ai-news view' first to see available articles",
		},
		{
			name: "article not found",
			args: []string{"5"},
			setupState: func() *state.ViewState {
				return &state.ViewState{
					Articles: map[string]state.ArticleRef{
						"1": {ID: 1, URL: "https://example.com", Title: "Test Article"},
					},
				}
			},
			expectedError: "article 5 not found in last view - available articles: run 'ai-news view' to see current list",
		},
		{
			name: "article has no URL",
			args: []string{"1"},
			setupState: func() *state.ViewState {
				return &state.ViewState{
					Articles: map[string]state.ArticleRef{
						"1": {ID: 1, URL: "", Title: "Test Article"},
					},
				}
			},
			expectedError: "article 1 has no URL available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestState(t, tt.setupState())

			cmd := &cobra.Command{}
			cmd.AddCommand(readCmd)

			var stdout, stderr bytes.Buffer
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)

			args := append([]string{"read"}, tt.args...)
			cmd.SetArgs(args)

			err := cmd.Execute()

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				if tt.expectedOutput != "" {
					assert.Contains(t, stdout.String(), tt.expectedOutput)
				}
			}
		})
	}
}

func TestReadCommandSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "https://example.com/article")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Test Article\n\nThis is test content."))
	}))
	defer server.Close()

	originalEndpoint := article.SetJinaEndpointForTesting(server.URL + "/%s")
	defer article.SetJinaEndpointForTesting(originalEndpoint)

	vs := &state.ViewState{
		Articles: map[string]state.ArticleRef{
			"1": {
				ID:    1,
				URL:   "https://example.com/article",
				Title: "Test Article",
			},
		},
	}
	setupTestState(t, vs)

	cmd := &cobra.Command{}
	cmd.AddCommand(readCmd)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"read", "1", "--no-style"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Reading: Test Article")
	assert.Contains(t, output, "# Test Article")
	assert.Contains(t, output, "This is test content.")
}

func TestReadCommandWithCache(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Cached Article\n\nThis is cached content."))
	}))
	defer server.Close()

	originalEndpoint := article.SetJinaEndpointForTesting(server.URL + "/%s")
	defer article.SetJinaEndpointForTesting(originalEndpoint)

	now := time.Now()
	vs := &state.ViewState{
		Articles: map[string]state.ArticleRef{
			"1": {
				ID:               1,
				URL:              "https://example.com/cached",
				Title:            "Cached Article",
				Content:          "# Cached Article\n\nThis is cached content.",
				ContentFetchedAt: &now,
			},
		},
	}
	setupTestState(t, vs)

	cmd := &cobra.Command{}
	cmd.AddCommand(readCmd)

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"read", "1", "--no-style"})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, 0, requestCount, "Should use cached content without making HTTP request")
	assert.Contains(t, stdout.String(), "This is cached content.")
}

func TestReadCommandNoCache(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		assert.Contains(t, r.URL.RawQuery, "no-cache=true")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Fresh Article\n\nThis is fresh content."))
	}))
	defer server.Close()

	originalEndpoint := article.SetJinaEndpointForTesting(server.URL + "/%s")
	defer article.SetJinaEndpointForTesting(originalEndpoint)

	now := time.Now()
	vs := &state.ViewState{
		Articles: map[string]state.ArticleRef{
			"1": {
				ID:               1,
				URL:              "https://example.com/fresh",
				Title:            "Fresh Article",
				Content:          "# Old Content",
				ContentFetchedAt: &now,
			},
		},
	}
	setupTestState(t, vs)

	cmd := &cobra.Command{}
	cmd.AddCommand(readCmd)

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"read", "1", "--no-cache", "--no-style"})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, 1, requestCount, "Should make HTTP request despite cached content")
	assert.Contains(t, stdout.String(), "This is fresh content.")
}

func TestReadCommandHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	originalEndpoint := article.SetJinaEndpointForTesting(server.URL + "/%s")
	defer article.SetJinaEndpointForTesting(originalEndpoint)

	vs := &state.ViewState{
		Articles: map[string]state.ArticleRef{
			"1": {
				ID:    1,
				URL:   "https://example.com/notfound",
				Title: "Not Found Article",
			},
		},
	}
	setupTestState(t, vs)

	cmd := &cobra.Command{}
	cmd.AddCommand(readCmd)

	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"read", "1"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "jina reader API returned 404")
}

func setupTestState(t *testing.T, vs *state.ViewState) {
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, ".ai-news-state.json")

	originalPathFunc := state.GetPathFunc()
	state.SetPathFunc(func() (string, error) {
		return statePath, nil
	})

	t.Cleanup(func() {
		state.SetPathFunc(originalPathFunc)
	})

	if vs != nil {
		err := state.Save(vs)
		require.NoError(t, err)
	}
}

func TestGetArticleContent(t *testing.T) {
	ctx := context.Background()

	t.Run("uses cached content when fresh", func(t *testing.T) {
		now := time.Now()
		ref := state.ArticleRef{
			Content:          "cached content",
			ContentFetchedAt: &now,
		}

		content, err := getArticleContent(ctx, ref, false)
		require.NoError(t, err)
		assert.Equal(t, "cached content", content)
	})

	t.Run("fetches fresh content when cache is old", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("fresh content"))
		}))
		defer server.Close()

		originalEndpoint := article.SetJinaEndpointForTesting(server.URL + "/%s")
		defer article.SetJinaEndpointForTesting(originalEndpoint)

		oldTime := time.Now().Add(-25 * time.Hour)
		ref := state.ArticleRef{
			URL:              "https://example.com",
			Content:          "old content",
			ContentFetchedAt: &oldTime,
		}

		content, err := getArticleContent(ctx, ref, false)
		require.NoError(t, err)
		assert.Equal(t, "fresh content", content)
	})

	t.Run("forces fresh fetch with no-cache", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("forced fresh content"))
		}))
		defer server.Close()

		originalEndpoint := article.SetJinaEndpointForTesting(server.URL + "/%s")
		defer article.SetJinaEndpointForTesting(originalEndpoint)

		now := time.Now()
		ref := state.ArticleRef{
			URL:              "https://example.com",
			Content:          "cached content",
			ContentFetchedAt: &now,
		}

		content, err := getArticleContent(ctx, ref, true)
		require.NoError(t, err)
		assert.Equal(t, "forced fresh content", content)
	})
}
