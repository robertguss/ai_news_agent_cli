package cmd

import (
	"bytes"
	"database/sql"
	"errors"
	"testing"

	"github.com/robertguss/ai-news-agent-cli/internal/config"
	"github.com/robertguss/ai-news-agent-cli/internal/database"
	"github.com/robertguss/ai-news-agent-cli/internal/fetcher"
	"github.com/stretchr/testify/assert"
)

func TestFetchCmd_ConfigLoadError(t *testing.T) {
	originalLoadCfg := loadCfg
	loadCfg = func(configPath string) (*config.Config, error) {
		return nil, errors.New("config load failed")
	}
	defer func() { loadCfg = originalLoadCfg }()

	cmd := NewRootCmd()
	cmd.AddCommand(fetchCmd)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"fetch", "--config", "nonexistent.yaml"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "load config: config load failed")
}

func TestFetchCmd_DatabaseOpenError(t *testing.T) {
	originalLoadCfg := loadCfg
	originalOpenDB := openDB

	loadCfg = func(configPath string) (*config.Config, error) {
		return &config.Config{
			DSN: "invalid-dsn",
			Sources: []fetcher.Source{
				{Name: "Test", URL: "http://example.com", Type: "rss", Priority: 1},
			},
		}, nil
	}
	openDB = func(dataSource string) (*sql.DB, *database.Queries, error) {
		return nil, nil, errors.New("database connection failed")
	}

	defer func() {
		loadCfg = originalLoadCfg
		openDB = originalOpenDB
	}()

	cmd := NewRootCmd()
	cmd.AddCommand(fetchCmd)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"fetch"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "open db: database connection failed")
}

func TestFetchCmd_SchemaInitError(t *testing.T) {
	originalLoadCfg := loadCfg
	originalOpenDB := openDB
	originalInitDB := initDB

	loadCfg = func(configPath string) (*config.Config, error) {
		return &config.Config{
			DSN: ":memory:",
			Sources: []fetcher.Source{
				{Name: "Test", URL: "http://example.com", Type: "rss", Priority: 1},
			},
		}, nil
	}
	openDB = func(dataSource string) (*sql.DB, *database.Queries, error) {
		db, err := sql.Open("sqlite", ":memory:")
		return db, database.New(db), err
	}
	initDB = func(db *sql.DB) error {
		return errors.New("schema init failed")
	}

	defer func() {
		loadCfg = originalLoadCfg
		openDB = originalOpenDB
		initDB = originalInitDB
	}()

	cmd := NewRootCmd()
	cmd.AddCommand(fetchCmd)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"fetch"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "init schema: schema init failed")
}

func TestFetchCmd_EmptySourcesList(t *testing.T) {
	originalLoadCfg := loadCfg
	originalOpenDB := openDB

	loadCfg = func(configPath string) (*config.Config, error) {
		return &config.Config{
			DSN:     ":memory:",
			Sources: []fetcher.Source{},
		}, nil
	}
	openDB = func(dataSource string) (*sql.DB, *database.Queries, error) {
		db, err := sql.Open("sqlite", ":memory:")
		return db, database.New(db), err
	}

	defer func() {
		loadCfg = originalLoadCfg
		openDB = originalOpenDB
	}()

	cmd := NewRootCmd()
	cmd.AddCommand(fetchCmd)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"fetch"})

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Added 0 new articles from 0 sources")
}
