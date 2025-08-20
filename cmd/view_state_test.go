package cmd

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/robertguss/ai-news-agent-cli/internal/database"
	"github.com/robertguss/ai-news-agent-cli/internal/state"
	"github.com/robertguss/ai-news-agent-cli/internal/tui"
	"github.com/stretchr/testify/assert"
)

func TestViewCommand_SavesStateAfterDisplay(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	statePath := filepath.Join(tmpDir, ".ai-news-state.json")

	originalPathFunc := state.GetPathFunc()
	state.SetPathFunc(func() (string, error) {
		return statePath, nil
	})
	defer func() { state.SetPathFunc(originalPathFunc) }()

	db, err := sql.Open("sqlite3", dbPath)
	assert.NoError(t, err)
	defer db.Close()

	err = database.InitSchema(db)
	assert.NoError(t, err)

	q := database.New(db)

	article1, err := q.CreateArticle(context.Background(), database.CreateArticleParams{
		Title:        sql.NullString{String: "Test Article 1", Valid: true},
		Url:          sql.NullString{String: "https://example.com/1", Valid: true},
		SourceName:   sql.NullString{String: "Test Source", Valid: true},
		Summary:      sql.NullString{String: "Test summary 1", Valid: true},
		Status:       sql.NullString{String: "unread", Valid: true},
		StoryGroupID: sql.NullString{String: "story-1", Valid: true},
	})
	assert.NoError(t, err)

	article2, err := q.CreateArticle(context.Background(), database.CreateArticleParams{
		Title:        sql.NullString{String: "Test Article 2", Valid: true},
		Url:          sql.NullString{String: "https://example.com/2", Valid: true},
		SourceName:   sql.NullString{String: "Test Source", Valid: true},
		Summary:      sql.NullString{String: "Test summary 2", Valid: true},
		Status:       sql.NullString{String: "unread", Valid: true},
		StoryGroupID: sql.NullString{String: "story-2", Valid: true},
	})
	assert.NoError(t, err)

	shouldUseTUIFunc = func() bool { return false }
	defer func() { shouldUseTUIFunc = tui.ShouldUseTUI }()

	opts := ViewOptions{All: true}
	err = runLegacyView(nil, dbPath, opts)
	assert.NoError(t, err)

	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Fatal("State file was not created")
	}

	loadedState, err := state.Load()
	assert.NoError(t, err)
	assert.NotNil(t, loadedState)

	assert.Len(t, loadedState.Articles, 2)

	article1State, exists := loadedState.Articles["1"]
	assert.True(t, exists)
	assert.Equal(t, article1.ID, article1State.ID)
	assert.Equal(t, "https://example.com/1", article1State.URL)
	assert.Equal(t, "Test Article 1", article1State.Title)
	assert.Equal(t, "story-1", article1State.StoryGroupID)

	article2State, exists := loadedState.Articles["2"]
	assert.True(t, exists)
	assert.Equal(t, article2.ID, article2State.ID)
	assert.Equal(t, "https://example.com/2", article2State.URL)
	assert.Equal(t, "Test Article 2", article2State.Title)
	assert.Equal(t, "story-2", article2State.StoryGroupID)

	assert.True(t, time.Since(loadedState.Timestamp) < time.Minute)
}

func TestViewCommand_DoesNotSaveStateWhenNoArticles(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "empty.db")
	statePath := filepath.Join(tmpDir, ".ai-news-state.json")

	originalPathFunc := state.GetPathFunc()
	state.SetPathFunc(func() (string, error) {
		return statePath, nil
	})
	defer func() { state.SetPathFunc(originalPathFunc) }()

	db, err := sql.Open("sqlite3", dbPath)
	assert.NoError(t, err)
	defer db.Close()

	err = database.InitSchema(db)
	assert.NoError(t, err)

	shouldUseTUIFunc = func() bool { return false }
	defer func() { shouldUseTUIFunc = tui.ShouldUseTUI }()

	opts := ViewOptions{All: true}
	err = runLegacyView(nil, dbPath, opts)
	assert.NoError(t, err)

	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Error("State file should not be created when no articles are displayed")
	}
}
