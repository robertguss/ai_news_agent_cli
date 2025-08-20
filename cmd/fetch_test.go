package cmd

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestFetchCmd_Integration_Success(t *testing.T) {
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

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `dsn: "` + dbPath + `"
sources:
  - name: "Test Source"
    url: "` + server.URL + `"
    type: "rss"
    priority: 1`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	cmd := NewRootCmd()
	cmd.AddCommand(fetchCmd)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"fetch", "--config", configPath})

	err = cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Added 2 new articles")
	assert.Contains(t, output, "from 1 sources")

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM articles").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	buf.Reset()
	cmd.SetArgs([]string{"fetch", "--config", configPath})
	err = cmd.Execute()
	assert.NoError(t, err)

	output = buf.String()
	assert.Contains(t, output, "Added 0 new articles")
	assert.Contains(t, output, "from 1 sources")
}

func TestFetchCmd_Integration_NetworkError(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `dsn: "` + dbPath + `"
sources:
  - name: "Bad Source"
    url: "http://nonexistent-domain-12345.com/feed.xml"
    type: "rss"
    priority: 1
  - name: "Another Bad Source"
    url: "http://another-nonexistent-domain-12345.com/feed.xml"
    type: "rss"
    priority: 1`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	cmd := NewRootCmd()
	cmd.AddCommand(fetchCmd)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"fetch", "--config", configPath})

	err = cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Added 0 new articles")
	assert.Contains(t, output, "2 errors occurred")
}

func TestFetchCmd_Integration_PartialSuccess(t *testing.T) {
	rssContent := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
    <channel>
        <title>Test Feed</title>
        <item>
            <title>Working Article</title>
            <link>https://example.com/working</link>
            <pubDate>Wed, 13 Aug 2025 20:28:20 +0000</pubDate>
        </item>
    </channel>
</rss>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(rssContent))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `dsn: "` + dbPath + `"
sources:
  - name: "Working Source"
    url: "` + server.URL + `"
    type: "rss"
    priority: 1
  - name: "Broken Source"
    url: "http://nonexistent-domain-12345.com/feed.xml"
    type: "rss"
    priority: 1`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	cmd := NewRootCmd()
	cmd.AddCommand(fetchCmd)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"fetch", "--config", configPath})

	err = cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Added 1 new articles")
	assert.Contains(t, output, "from 2 sources")
	assert.Contains(t, output, "1 errors occurred")

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM articles").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
