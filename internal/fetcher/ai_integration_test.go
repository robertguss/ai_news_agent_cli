package fetcher

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/robertguss/ai-news-agent-cli/internal/ai/processor"
	"github.com/robertguss/ai-news-agent-cli/internal/ai/processor/mocks"
	"github.com/robertguss/ai-news-agent-cli/internal/database"
	"github.com/robertguss/ai-news-agent-cli/internal/scraper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestStoreArticlesWithAI_Success(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	defer db.Close()

	err = database.InitSchema(db)
	require.NoError(t, err)

	queries := database.New(db)

	mockScraper := scraper.NewMockScraper("scraped article content", nil)
	mockAI := new(mocks.AIProcessor)
	mockAI.On("AnalyzeContent", "scraped article content").Return(&processor.AnalysisResult{
		Summary: "AI generated summary",
	}, nil)

	deps := PipelineDeps{
		Scraper: mockScraper,
		AI:      mockAI,
		Queries: queries,
	}

	articles := []Article{
		{
			Title:         "Test Article",
			Link:          "https://example.com/test",
			PublishedDate: time.Now(),
		},
	}

	source := Source{
		Name: "Test Source",
	}

	ctx := context.Background()
	stored, err := StoreArticlesWithAI(ctx, deps, articles, source)

	require.NoError(t, err)
	assert.Equal(t, 1, stored)

	var summary sql.NullString
	err = db.QueryRow("SELECT summary FROM articles WHERE url = ?", "https://example.com/test").Scan(&summary)
	require.NoError(t, err)
	assert.True(t, summary.Valid)
	assert.Equal(t, "AI generated summary", summary.String)

	mockAI.AssertExpectations(t)
}

func TestStoreArticlesWithAI_ScraperError(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	defer db.Close()

	err = database.InitSchema(db)
	require.NoError(t, err)

	queries := database.New(db)

	mockScraper := scraper.NewMockScraper("", assert.AnError)
	mockAI := new(mocks.AIProcessor)

	deps := PipelineDeps{
		Scraper: mockScraper,
		AI:      mockAI,
		Queries: queries,
	}

	articles := []Article{
		{
			Title:         "Test Article",
			Link:          "https://example.com/test",
			PublishedDate: time.Now(),
		},
	}

	source := Source{
		Name: "Test Source",
	}

	ctx := context.Background()
	stored, err := StoreArticlesWithAI(ctx, deps, articles, source)

	require.NoError(t, err)
	assert.Equal(t, 1, stored)

	var summary sql.NullString
	err = db.QueryRow("SELECT summary FROM articles WHERE url = ?", "https://example.com/test").Scan(&summary)
	require.NoError(t, err)
	assert.False(t, summary.Valid)

	mockAI.AssertNotCalled(t, "AnalyzeContent", mock.Anything)
}

func TestStoreArticlesWithAI_AIError(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	defer db.Close()

	err = database.InitSchema(db)
	require.NoError(t, err)

	queries := database.New(db)

	mockScraper := scraper.NewMockScraper("scraped content", nil)
	mockAI := new(mocks.AIProcessor)
	mockAI.On("AnalyzeContent", "scraped content").Return(nil, assert.AnError)

	deps := PipelineDeps{
		Scraper: mockScraper,
		AI:      mockAI,
		Queries: queries,
	}

	articles := []Article{
		{
			Title:         "Test Article",
			Link:          "https://example.com/test",
			PublishedDate: time.Now(),
		},
	}

	source := Source{
		Name: "Test Source",
	}

	ctx := context.Background()
	stored, err := StoreArticlesWithAI(ctx, deps, articles, source)

	require.NoError(t, err)
	assert.Equal(t, 1, stored)

	var summary sql.NullString
	err = db.QueryRow("SELECT summary FROM articles WHERE url = ?", "https://example.com/test").Scan(&summary)
	require.NoError(t, err)
	assert.False(t, summary.Valid)

	mockAI.AssertExpectations(t)
}
