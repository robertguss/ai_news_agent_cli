package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStateIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, ".ai-news-state.json")

	originalPathFunc := pathFunc
	pathFunc = func() (string, error) {
		return statePath, nil
	}
	defer func() { pathFunc = originalPathFunc }()

	timestamp := time.Date(2025, 8, 19, 21, 30, 0, 0, time.UTC)
	testState := &ViewState{
		Timestamp: timestamp,
		Articles: map[string]ArticleRef{
			"1": {
				ID:           123,
				URL:          "https://example.com/article1",
				Title:        "Test Article 1",
				StoryGroupID: "group1",
			},
			"2": {
				ID:           124,
				URL:          "https://example.com/article2",
				Title:        "Test Article 2",
				StoryGroupID: "group2",
			},
		},
	}

	err := Save(testState)
	if err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Fatal("State file was not created")
	}

	loadedState, err := Load()
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	if len(loadedState.Articles) != 2 {
		t.Errorf("Expected 2 articles, got %d", len(loadedState.Articles))
	}

	article1, exists := loadedState.Articles["1"]
	if !exists {
		t.Fatal("Article 1 not found")
	}

	if article1.ID != 123 {
		t.Errorf("Article 1 ID mismatch: got %d, want 123", article1.ID)
	}

	if article1.Title != "Test Article 1" {
		t.Errorf("Article 1 title mismatch: got %s, want Test Article 1", article1.Title)
	}

	t.Logf("State file created at: %s", statePath)
	t.Logf("State contains %d articles", len(loadedState.Articles))
}
