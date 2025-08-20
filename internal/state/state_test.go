package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, ".ai-news-state.json")

	originalPathFunc := pathFunc
	pathFunc = func() (string, error) {
		return statePath, nil
	}
	defer func() { pathFunc = originalPathFunc }()

	timestamp := time.Date(2025, 8, 19, 21, 30, 0, 0, time.UTC)
	originalState := &ViewState{
		Timestamp: timestamp,
		Articles: map[string]ArticleRef{
			"1": {
				ID:           123,
				URL:          "https://example.com/article1",
				Title:        "Article Title",
				StoryGroupID: "group1",
			},
			"2": {
				ID:           124,
				URL:          "https://example.com/article2",
				Title:        "Another Article",
				StoryGroupID: "group2",
			},
		},
	}

	err := Save(originalState)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loadedState, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !loadedState.Timestamp.Equal(originalState.Timestamp) {
		t.Errorf("Timestamp mismatch: got %v, want %v", loadedState.Timestamp, originalState.Timestamp)
	}

	if len(loadedState.Articles) != len(originalState.Articles) {
		t.Errorf("Articles count mismatch: got %d, want %d", len(loadedState.Articles), len(originalState.Articles))
	}

	for key, expectedArticle := range originalState.Articles {
		actualArticle, exists := loadedState.Articles[key]
		if !exists {
			t.Errorf("Article %s not found in loaded state", key)
			continue
		}

		if actualArticle.ID != expectedArticle.ID {
			t.Errorf("Article %s ID mismatch: got %d, want %d", key, actualArticle.ID, expectedArticle.ID)
		}
		if actualArticle.URL != expectedArticle.URL {
			t.Errorf("Article %s URL mismatch: got %s, want %s", key, actualArticle.URL, expectedArticle.URL)
		}
		if actualArticle.Title != expectedArticle.Title {
			t.Errorf("Article %s Title mismatch: got %s, want %s", key, actualArticle.Title, expectedArticle.Title)
		}
		if actualArticle.StoryGroupID != expectedArticle.StoryGroupID {
			t.Errorf("Article %s StoryGroupID mismatch: got %s, want %s", key, actualArticle.StoryGroupID, expectedArticle.StoryGroupID)
		}
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "nonexistent.json")

	originalPathFunc := pathFunc
	pathFunc = func() (string, error) {
		return statePath, nil
	}
	defer func() { pathFunc = originalPathFunc }()

	state, err := Load()
	if err != nil {
		t.Fatalf("Load should not fail for non-existent file: %v", err)
	}

	if state == nil {
		t.Fatal("Load should return empty state, not nil")
	}

	if len(state.Articles) != 0 {
		t.Errorf("Empty state should have 0 articles, got %d", len(state.Articles))
	}
}

func TestLoadCorruptedFile(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, ".ai-news-state.json")

	originalPathFunc := pathFunc
	pathFunc = func() (string, error) {
		return statePath, nil
	}
	defer func() { pathFunc = originalPathFunc }()

	err := os.WriteFile(statePath, []byte("invalid json"), 0o600)
	if err != nil {
		t.Fatalf("Failed to write corrupted file: %v", err)
	}

	_, err = Load()
	if err == nil {
		t.Error("Load should fail for corrupted JSON file")
	}
}

func TestSaveFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, ".ai-news-state.json")

	originalPathFunc := pathFunc
	pathFunc = func() (string, error) {
		return statePath, nil
	}
	defer func() { pathFunc = originalPathFunc }()

	state := &ViewState{
		Timestamp: time.Now().UTC(),
		Articles:  map[string]ArticleRef{},
	}

	err := Save(state)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	info, err := os.Stat(statePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	expectedMode := os.FileMode(0o600)
	if info.Mode().Perm() != expectedMode {
		t.Errorf("File permissions mismatch: got %o, want %o", info.Mode().Perm(), expectedMode)
	}
}

func TestJSONFormat(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, ".ai-news-state.json")

	originalPathFunc := pathFunc
	pathFunc = func() (string, error) {
		return statePath, nil
	}
	defer func() { pathFunc = originalPathFunc }()

	timestamp := time.Date(2025, 8, 19, 21, 30, 0, 0, time.UTC)
	state := &ViewState{
		Timestamp: timestamp,
		Articles: map[string]ArticleRef{
			"1": {
				ID:           123,
				URL:          "https://example.com/article1",
				Title:        "Article Title",
				StoryGroupID: "group1",
			},
		},
	}

	err := Save(state)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	content, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var parsed map[string]interface{}
	err = json.Unmarshal(content, &parsed)
	if err != nil {
		t.Fatalf("Generated JSON is invalid: %v", err)
	}

	articles, ok := parsed["articles"].(map[string]interface{})
	if !ok {
		t.Fatal("Articles field should be an object")
	}

	article1, ok := articles["1"].(map[string]interface{})
	if !ok {
		t.Fatal("Article 1 should be an object")
	}

	if article1["id"].(float64) != 123 {
		t.Errorf("Article ID mismatch: got %v, want 123", article1["id"])
	}

	if article1["url"].(string) != "https://example.com/article1" {
		t.Errorf("Article URL mismatch: got %v, want https://example.com/article1", article1["url"])
	}
}

func TestEmptyArticles(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, ".ai-news-state.json")

	originalPathFunc := pathFunc
	pathFunc = func() (string, error) {
		return statePath, nil
	}
	defer func() { pathFunc = originalPathFunc }()

	state := &ViewState{
		Timestamp: time.Now().UTC(),
		Articles:  map[string]ArticleRef{},
	}

	err := Save(state)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loadedState, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(loadedState.Articles) != 0 {
		t.Errorf("Empty articles should remain empty, got %d articles", len(loadedState.Articles))
	}
}
