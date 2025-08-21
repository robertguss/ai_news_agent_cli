package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_FromConfigsDirectory(t *testing.T) {
	tempDir := t.TempDir()
	configsDir := filepath.Join(tempDir, "configs")
	err := os.MkdirAll(configsDir, 0755)
	require.NoError(t, err)

	configPath := filepath.Join(configsDir, "config.yaml")

	configContent := `sources:
  - name: "Configs Dir Source"
    url: "https://example.com/configs-feed.xml"
    type: "rss"
    priority: 3`

	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	config, err := Load()
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Len(t, config.Sources, 1)
	assert.Equal(t, "Configs Dir Source", config.Sources[0].Name)
	assert.Equal(t, "https://example.com/configs-feed.xml", config.Sources[0].URL)
	assert.Equal(t, "rss", config.Sources[0].Type)
	assert.Equal(t, 3, config.Sources[0].Priority)
}
