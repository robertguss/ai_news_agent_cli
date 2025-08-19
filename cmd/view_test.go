package cmd

import (
        "bytes"
        "context"
        "database/sql"
        "os"
        "path/filepath"
        "testing"

        "github.com/robertguss/ai-news-agent-cli/internal/database"
        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
        db, err := sql.Open("sqlite", ":memory:")
        require.NoError(t, err)
        
        err = database.InitSchema(db)
        require.NoError(t, err)
        
        return db, func() {
                db.Close()
        }
}

func insertTestArticle(db *sql.DB, title, sourceName string) {
        var titleVal, sourceVal sql.NullString
        
        if title != "" {
                titleVal = sql.NullString{String: title, Valid: true}
        }
        if sourceName != "" {
                sourceVal = sql.NullString{String: sourceName, Valid: true}
        }
        
        q := database.New(db)
        _, err := q.CreateArticle(context.Background(), database.CreateArticleParams{
                Title:      titleVal,
                SourceName: sourceVal,
        })
        if err != nil {
                panic(err)
        }
}

func executeViewCommand(args ...string) (string, error) {
        cmd := NewRootCmd()
        cmd.AddCommand(viewCmd)
        
        var buf bytes.Buffer
        cmd.SetOut(&buf)
        cmd.SetErr(&buf)
        cmd.SetArgs(args)
        
        err := cmd.Execute()
        return buf.String(), err
}

func TestViewCmd_EmptyDatabase(t *testing.T) {
        db, cleanup := setupTestDB(t)
        defer cleanup()
        
        originalOpen := databaseOpen
        databaseOpen = func(dataSource string) (*sql.DB, *database.Queries, error) {
                return db, database.New(db), nil
        }
        defer func() { databaseOpen = originalOpen }()
        
        output, err := executeViewCommand("view")
        
        assert.NoError(t, err)
        assert.Equal(t, "No articles found.\n", output)
}

func TestViewCmd_SingleArticle(t *testing.T) {
        db, cleanup := setupTestDB(t)
        defer cleanup()
        
        insertTestArticle(db, "Test Title", "Test Source")
        
        originalOpen := databaseOpen
        databaseOpen = func(dataSource string) (*sql.DB, *database.Queries, error) {
                return db, database.New(db), nil
        }
        defer func() { databaseOpen = originalOpen }()
        
        output, err := executeViewCommand("view")
        
        assert.NoError(t, err)
        assert.Equal(t, "Test Title - Test Source\n", output)
}

func TestViewCmd_MultipleArticles(t *testing.T) {
        db, cleanup := setupTestDB(t)
        defer cleanup()
        
        insertTestArticle(db, "First Article", "BBC")
        insertTestArticle(db, "Second Article", "Reuters")
        insertTestArticle(db, "Third Article", "CNN")
        
        originalOpen := databaseOpen
        databaseOpen = func(dataSource string) (*sql.DB, *database.Queries, error) {
                return db, database.New(db), nil
        }
        defer func() { databaseOpen = originalOpen }()
        
        output, err := executeViewCommand("view")
        
        assert.NoError(t, err)
        expected := "First Article - BBC\nSecond Article - Reuters\nThird Article - CNN\n"
        assert.Equal(t, expected, output)
}

func TestViewCmd_NullValues(t *testing.T) {
        db, cleanup := setupTestDB(t)
        defer cleanup()
        
        insertTestArticle(db, "", "Test Source")
        insertTestArticle(db, "Test Title", "")
        insertTestArticle(db, "", "")
        
        originalOpen := databaseOpen
        databaseOpen = func(dataSource string) (*sql.DB, *database.Queries, error) {
                return db, database.New(db), nil
        }
        defer func() { databaseOpen = originalOpen }()
        
        output, err := executeViewCommand("view")
        
        assert.NoError(t, err)
        expected := "(no title) - Test Source\nTest Title - (no source)\n(no title) - (no source)\n"
        assert.Equal(t, expected, output)
}

func TestViewCmd_DatabaseFlag(t *testing.T) {
        tmpDir := t.TempDir()
        dbPath := filepath.Join(tmpDir, "test.db")
        
        db, err := sql.Open("sqlite", dbPath)
        require.NoError(t, err)
        defer db.Close()
        
        err = database.InitSchema(db)
        require.NoError(t, err)
        
        insertTestArticle(db, "Flag Test", "Test Source")
        
        output, err := executeViewCommand("view", "--db", dbPath)
        
        assert.NoError(t, err)
        assert.Equal(t, "Flag Test - Test Source\n", output)
}

func TestViewCmd_DatabaseConnectionError(t *testing.T) {
        output, err := executeViewCommand("view", "--db", "/invalid/path/db.sqlite")
        
        assert.Error(t, err)
        assert.Contains(t, output, "Error:")
}

func TestViewCmd_DatabaseAutoCreation(t *testing.T) {
        tmpDir := t.TempDir()
        dbPath := filepath.Join(tmpDir, "new_test.db")
        
        _, err := os.Stat(dbPath)
        assert.True(t, os.IsNotExist(err))
        
        output, err := executeViewCommand("view", "--db", dbPath)
        
        assert.NoError(t, err)
        assert.Equal(t, "No articles found.\n", output)
        
        _, err = os.Stat(dbPath)
        assert.NoError(t, err)
}

func TestViewCmd_Integration(t *testing.T) {
        tmpDir := t.TempDir()
        dbPath := filepath.Join(tmpDir, "integration.db")
        
        db, err := sql.Open("sqlite", dbPath)
        require.NoError(t, err)
        
        err = database.InitSchema(db)
        require.NoError(t, err)
        
        insertTestArticle(db, "Integration Test", "Integration Source")
        db.Close()
        
        output, err := executeViewCommand("view", "--db", dbPath)
        
        assert.NoError(t, err)
        assert.Equal(t, "Integration Test - Integration Source\n", output)
}
