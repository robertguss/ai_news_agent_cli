package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Success(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	
	configContent := `sources:
  - name: "Test Source"
    url: "https://example.com/feed.xml"
    type: "rss"
    priority: 1
  - name: "Another Source"
    url: "https://example.com/feed2.xml"
    type: "rss"
    priority: 2`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	config, err := Load()
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Len(t, config.Sources, 2)
	assert.Equal(t, "Test Source", config.Sources[0].Name)
	assert.Equal(t, "https://example.com/feed.xml", config.Sources[0].URL)
	assert.Equal(t, "rss", config.Sources[0].Type)
	assert.Equal(t, 1, config.Sources[0].Priority)
	assert.Equal(t, "Another Source", config.Sources[1].Name)
	assert.Equal(t, 2, config.Sources[1].Priority)
}

func TestLoad_MissingFile(t *testing.T) {
	tempDir := t.TempDir()
	
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	config, err := Load()
	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestLoad_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	
	invalidYAML := `sources:
  - name: "Test Source"
    url: "https://example.com/feed.xml"
    type: "rss"
    priority: invalid_number`

	err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	config, err := Load()
	assert.Error(t, err)
	assert.Nil(t, config)
}
