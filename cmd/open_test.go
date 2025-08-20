package cmd

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/robertguss/ai-news-agent-cli/internal/browserutil"
	"github.com/robertguss/ai-news-agent-cli/internal/state"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenCmd_Success(t *testing.T) {
	tmpHome := t.TempDir()
	originalPathFunc := state.GetPathFunc()
	state.SetPathFunc(func() (string, error) {
		return filepath.Join(tmpHome, ".ai-news-state.json"), nil
	})
	t.Cleanup(func() {
		state.SetPathFunc(originalPathFunc)
	})

	vs := &state.ViewState{
		Timestamp: time.Now(),
		Articles: map[string]state.ArticleRef{
			"1": {
				ID:    1,
				URL:   "https://example.com/article1",
				Title: "Test Article 1",
			},
			"2": {
				ID:    2,
				URL:   "https://example.com/article2",
				Title: "Test Article 2",
			},
		},
	}
	require.NoError(t, state.Save(vs))

	var openedURL string
	originalOpenURL := browserutil.OpenURL
	browserutil.OpenURL = func(url string) error {
		openedURL = url
		return nil
	}
	t.Cleanup(func() {
		browserutil.OpenURL = originalOpenURL
	})

	cmd := &cobra.Command{}
	cmd.AddCommand(openCmd)

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"open", "1"})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, "https://example.com/article1", openedURL)
	outputStr := output.String()
	assert.Contains(t, outputStr, "Opening: Test Article 1")
	assert.Contains(t, outputStr, "https://example.com/article1")
}

func TestOpenCmd_InvalidArticleNumber(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.AddCommand(openCmd)

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"open", "invalid"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid article number")
	assert.Contains(t, err.Error(), "must be a positive integer")
}

func TestOpenCmd_NoStateFile(t *testing.T) {
	tmpHome := t.TempDir()
	originalPathFunc := state.GetPathFunc()
	state.SetPathFunc(func() (string, error) {
		return filepath.Join(tmpHome, ".ai-news-state.json"), nil
	})
	t.Cleanup(func() {
		state.SetPathFunc(originalPathFunc)
	})

	cmd := &cobra.Command{}
	cmd.AddCommand(openCmd)

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"open", "1"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no viewed articles found")
	assert.Contains(t, err.Error(), "run 'ai-news view' first")
}

func TestOpenCmd_EmptyState(t *testing.T) {
	tmpHome := t.TempDir()
	originalPathFunc := state.GetPathFunc()
	state.SetPathFunc(func() (string, error) {
		return filepath.Join(tmpHome, ".ai-news-state.json"), nil
	})
	t.Cleanup(func() {
		state.SetPathFunc(originalPathFunc)
	})

	vs := &state.ViewState{
		Timestamp: time.Now(),
		Articles:  map[string]state.ArticleRef{},
	}
	require.NoError(t, state.Save(vs))

	cmd := &cobra.Command{}
	cmd.AddCommand(openCmd)

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"open", "1"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no viewed articles found")
	assert.Contains(t, err.Error(), "run 'ai-news view' first")
}

func TestOpenCmd_ArticleNotFound(t *testing.T) {
	tmpHome := t.TempDir()
	originalPathFunc := state.GetPathFunc()
	state.SetPathFunc(func() (string, error) {
		return filepath.Join(tmpHome, ".ai-news-state.json"), nil
	})
	t.Cleanup(func() {
		state.SetPathFunc(originalPathFunc)
	})

	vs := &state.ViewState{
		Timestamp: time.Now(),
		Articles: map[string]state.ArticleRef{
			"1": {
				ID:    1,
				URL:   "https://example.com/article1",
				Title: "Test Article 1",
			},
		},
	}
	require.NoError(t, state.Save(vs))

	cmd := &cobra.Command{}
	cmd.AddCommand(openCmd)

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"open", "5"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "article 5 not found in last view")
}

func TestOpenCmd_EmptyURL(t *testing.T) {
	tmpHome := t.TempDir()
	originalPathFunc := state.GetPathFunc()
	state.SetPathFunc(func() (string, error) {
		return filepath.Join(tmpHome, ".ai-news-state.json"), nil
	})
	t.Cleanup(func() {
		state.SetPathFunc(originalPathFunc)
	})

	vs := &state.ViewState{
		Timestamp: time.Now(),
		Articles: map[string]state.ArticleRef{
			"1": {
				ID:    1,
				URL:   "",
				Title: "Test Article 1",
			},
		},
	}
	require.NoError(t, state.Save(vs))

	cmd := &cobra.Command{}
	cmd.AddCommand(openCmd)

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"open", "1"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "article 1 has no URL available")
}

func TestOpenCmd_BrowserOpenFailure(t *testing.T) {
	tmpHome := t.TempDir()
	originalPathFunc := state.GetPathFunc()
	state.SetPathFunc(func() (string, error) {
		return filepath.Join(tmpHome, ".ai-news-state.json"), nil
	})
	t.Cleanup(func() {
		state.SetPathFunc(originalPathFunc)
	})

	vs := &state.ViewState{
		Timestamp: time.Now(),
		Articles: map[string]state.ArticleRef{
			"1": {
				ID:    1,
				URL:   "https://example.com/article1",
				Title: "Test Article 1",
			},
		},
	}
	require.NoError(t, state.Save(vs))

	originalOpenURL := browserutil.OpenURL
	browserutil.OpenURL = func(url string) error {
		return errors.New("browser failed to open")
	}
	t.Cleanup(func() {
		browserutil.OpenURL = originalOpenURL
	})

	cmd := &cobra.Command{}
	cmd.AddCommand(openCmd)

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"open", "1"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open browser")
	assert.Contains(t, err.Error(), "browser failed to open")
}

func TestOpenCmd_StateLoadFailure(t *testing.T) {
	originalPathFunc := state.GetPathFunc()
	state.SetPathFunc(func() (string, error) {
		return "", errors.New("path error")
	})
	t.Cleanup(func() {
		state.SetPathFunc(originalPathFunc)
	})

	cmd := &cobra.Command{}
	cmd.AddCommand(openCmd)

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"open", "1"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load view state")
}

func TestOpenCmd_WrongNumberOfArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"no args", []string{"open"}},
		{"too many args", []string{"open", "1", "2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.AddCommand(openCmd)

			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			require.Error(t, err)
		})
	}
}

func TestOpenCmd_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpHome := t.TempDir()
	originalPathFunc := state.GetPathFunc()
	state.SetPathFunc(func() (string, error) {
		return filepath.Join(tmpHome, ".ai-news-state.json"), nil
	})
	t.Cleanup(func() {
		state.SetPathFunc(originalPathFunc)
	})

	vs := &state.ViewState{
		Timestamp: time.Now(),
		Articles: map[string]state.ArticleRef{
			"1": {
				ID:    1,
				URL:   "https://example.com",
				Title: "Integration Test Article",
			},
		},
	}
	require.NoError(t, state.Save(vs))

	var openedURL string
	originalOpenURL := browserutil.OpenURL
	browserutil.OpenURL = func(url string) error {
		openedURL = url
		return nil
	}
	t.Cleanup(func() {
		browserutil.OpenURL = originalOpenURL
	})

	cmd := &cobra.Command{}
	cmd.AddCommand(openCmd)

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"open", "1"})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, "https://example.com", openedURL)
	outputStr := output.String()
	assert.Contains(t, outputStr, "Opening: Integration Test Article")
	assert.Contains(t, outputStr, "https://example.com")
}
