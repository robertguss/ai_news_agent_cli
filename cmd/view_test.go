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

func setupTestDB(t *testing.T) (*sql.DB, string, func()) {
        tmpDir := t.TempDir()
        dbPath := filepath.Join(tmpDir, "test.db")
        
        db, err := sql.Open("sqlite", dbPath)
        require.NoError(t, err)

        err = database.InitSchema(db)
        require.NoError(t, err)

        return db, dbPath, func() {
                db.Close()
        }
}

func insertTestArticle(db *sql.DB, title, sourceName string) {
        insertTestArticleWithDetails(db, title, sourceName, "unread", "", "", "")
}

func insertTestArticleWithDetails(db *sql.DB, title, sourceName, status, summary, topics, storyGroupID string) {
        var titleVal, sourceVal, summaryVal, storyGroupVal sql.NullString
        var topicsVal interface{}

        if title != "" {
                titleVal = sql.NullString{String: title, Valid: true}
        }
        if sourceName != "" {
                sourceVal = sql.NullString{String: sourceName, Valid: true}
        }
        if summary != "" {
                summaryVal = sql.NullString{String: summary, Valid: true}
        }
        if topics != "" {
                topicsVal = topics
        }
        if storyGroupID != "" {
                storyGroupVal = sql.NullString{String: storyGroupID, Valid: true}
        }

        q := database.New(db)
        _, err := q.CreateArticle(context.Background(), database.CreateArticleParams{
                Title:        titleVal,
                Url:          sql.NullString{String: "http://example.com/" + title + "-" + sourceName, Valid: true},
                SourceName:   sourceVal,
                Summary:      summaryVal,
                Topics:       topicsVal,
                Status:       sql.NullString{String: status, Valid: true},
                StoryGroupID: storyGroupVal,
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
        db, _, cleanup := setupTestDB(t)
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
        db, _, cleanup := setupTestDB(t)
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
        db, _, cleanup := setupTestDB(t)
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
        db, _, cleanup := setupTestDB(t)
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

func TestViewCmd_DefaultShowsOnlyUnreadAndMarksRead(t *testing.T) {
        db, dbPath, cleanup := setupTestDB(t)
        defer cleanup()

        insertTestArticleWithDetails(db, "Unread Article 1", "Google AI Blog", "unread", "Summary of unread article", `["Large Language Models", "AI"]`, "story-1")
        insertTestArticleWithDetails(db, "Read Article", "OpenAI Blog", "read", "Summary of read article", `["GPT", "AI"]`, "story-2")
        insertTestArticleWithDetails(db, "Unread Article 2", "Ars Technica AI", "unread", "Another unread summary", `["Machine Learning"]`, "story-3")

        output, err := executeViewCommand("view", "--db", dbPath)

        assert.NoError(t, err)
        assert.Contains(t, output, "Unread Article 1")
        assert.Contains(t, output, "Unread Article 2")
        assert.NotContains(t, output, "Read Article")

        q := database.New(db)
        articles, _ := q.ListArticles(context.Background())
        readCount := 0
        for _, article := range articles {
                if article.Status.String == "read" {
                        readCount++
                }
        }
        assert.Equal(t, 3, readCount)
}

func TestViewCmd_AllFlagShowsAllArticlesAndDoesNotMarkRead(t *testing.T) {
        db, dbPath, cleanup := setupTestDB(t)
        defer cleanup()

        insertTestArticleWithDetails(db, "Unread Article", "Google AI Blog", "unread", "Summary", `["AI"]`, "story-1")
        insertTestArticleWithDetails(db, "Read Article", "OpenAI Blog", "read", "Summary", `["GPT"]`, "story-2")

        output, err := executeViewCommand("view", "--all", "--db", dbPath)

        assert.NoError(t, err)
        assert.Contains(t, output, "Unread Article")
        assert.Contains(t, output, "Read Article")

        q := database.New(db)
        articles, _ := q.ListArticles(context.Background())
        unreadCount := 0
        for _, article := range articles {
                if article.Status.String == "unread" {
                        unreadCount++
                }
        }
        assert.Equal(t, 1, unreadCount)
}

func TestViewCmd_SourceFilterWorksCorrectly(t *testing.T) {
        db, dbPath, cleanup := setupTestDB(t)
        defer cleanup()

        insertTestArticleWithDetails(db, "Google Article", "Google AI Blog", "unread", "Summary", `["AI"]`, "story-1")
        insertTestArticleWithDetails(db, "OpenAI Article", "OpenAI Blog", "unread", "Summary", `["GPT"]`, "story-2")
        insertTestArticleWithDetails(db, "Ars Article", "Ars Technica AI", "unread", "Summary", `["Tech"]`, "story-3")

        output, err := executeViewCommand("view", "--source", "Google AI Blog", "--db", dbPath)

        assert.NoError(t, err)
        assert.Contains(t, output, "Google Article")
        assert.NotContains(t, output, "OpenAI Article")
        assert.NotContains(t, output, "Ars Article")
}

func TestViewCmd_TopicFilterWorksCorrectly(t *testing.T) {
        db, dbPath, cleanup := setupTestDB(t)
        defer cleanup()

        insertTestArticleWithDetails(db, "AI Article", "Google AI Blog", "unread", "Summary", `["Large Language Models", "AI"]`, "story-1")
        insertTestArticleWithDetails(db, "GPT Article", "OpenAI Blog", "unread", "Summary", `["GPT", "Transformers"]`, "story-2")
        insertTestArticleWithDetails(db, "ML Article", "Ars Technica AI", "unread", "Summary", `["Machine Learning", "AI"]`, "story-3")

        output, err := executeViewCommand("view", "--topic", "GPT", "--db", dbPath)

        assert.NoError(t, err)
        assert.Contains(t, output, "GPT Article")
        assert.NotContains(t, output, "AI Article")
        assert.NotContains(t, output, "ML Article")
}

func TestViewCmd_StyledOutputContainsExpectedElements(t *testing.T) {
        db, dbPath, cleanup := setupTestDB(t)
        defer cleanup()

        insertTestArticleWithDetails(db, "Google AI Announces Project Astra", "Google AI Blog", "unread", 
                "A new multimodal AI assistant was announced that can reason about video and audio in real-time", 
                `["Large Language Models", "Multimodal AI"]`, "story-1")

        output, err := executeViewCommand("view", "--db", dbPath)

        assert.NoError(t, err)
        assert.Contains(t, output, "[1]")
        assert.Contains(t, output, "Google AI Announces Project Astra")
        assert.Contains(t, output, "Google AI Blog")
        assert.Contains(t, output, "Tier 1")
        assert.Contains(t, output, "Summary:")
        assert.Contains(t, output, "Topics:")
        assert.Contains(t, output, "Large Language Models")
        assert.Contains(t, output, "Multimodal AI")
}

func TestViewCmd_StoryGroupingShowsDuplicates(t *testing.T) {
        db, dbPath, cleanup := setupTestDB(t)
        defer cleanup()

        insertTestArticleWithDetails(db, "Primary Article", "Google AI Blog", "unread", "Summary", `["AI"]`, "story-group-1")
        insertTestArticleWithDetails(db, "Duplicate Article 1", "The Verge", "unread", "Summary", `["AI"]`, "story-group-1")
        insertTestArticleWithDetails(db, "Duplicate Article 2", "Ars Technica", "unread", "Summary", `["AI"]`, "story-group-1")
        insertTestArticleWithDetails(db, "Different Story", "OpenAI Blog", "unread", "Summary", `["GPT"]`, "story-group-2")

        output, err := executeViewCommand("view", "--db", dbPath)

        assert.NoError(t, err)
        assert.Contains(t, output, "Primary Article")
        assert.Contains(t, output, "Also covered by:")
        assert.Contains(t, output, "The Verge")
        assert.Contains(t, output, "Ars Technica")
        assert.NotContains(t, output, "Duplicate Article 1")
        assert.NotContains(t, output, "Duplicate Article 2")
        assert.Contains(t, output, "Different Story")
}
