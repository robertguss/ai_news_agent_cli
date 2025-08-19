package database

import (
        "context"
        "database/sql"
        "testing"
        "time"

        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*sql.DB, *Queries) {
        db, queries, err := Open(":memory:")
        require.NoError(t, err)

        err = InitSchema(db)
        require.NoError(t, err)

        return db, queries
}

func TestCreateArticle(t *testing.T) {
        _, queries := setupTestDB(t)
        ctx := context.Background()

        params := CreateArticleParams{
                Title:         sql.NullString{String: "Test Article", Valid: true},
                Url:           sql.NullString{String: "https://example.com/test", Valid: true},
                SourceName:    sql.NullString{String: "Test Source", Valid: true},
                PublishedDate: sql.NullTime{Time: time.Now(), Valid: true},
                Summary:       sql.NullString{String: "Test summary", Valid: true},
                Entities:      `{"entities": ["test"]}`,
                ContentType:   sql.NullString{String: "article", Valid: true},
                Topics:        `{"topics": ["tech"]}`,
                Status:        sql.NullString{String: "unread", Valid: true},
                StoryGroupID:  sql.NullString{String: "group-1", Valid: true},
        }

        article, err := queries.CreateArticle(ctx, params)
        require.NoError(t, err)

        assert.Greater(t, article.ID, int64(0))
        assert.Equal(t, params.Title.String, article.Title.String)
        assert.Equal(t, params.Url.String, article.Url.String)
        assert.Equal(t, params.SourceName.String, article.SourceName.String)
        assert.Equal(t, params.Summary.String, article.Summary.String)
        assert.Equal(t, params.ContentType.String, article.ContentType.String)
        assert.Equal(t, params.Status.String, article.Status.String)
        assert.Equal(t, params.StoryGroupID.String, article.StoryGroupID.String)
}

func TestGetArticleByUrl(t *testing.T) {
        _, queries := setupTestDB(t)
        ctx := context.Background()

        testURL := "https://example.com/unique-test"

        params := CreateArticleParams{
                Title:         sql.NullString{String: "Test Article", Valid: true},
                Url:           sql.NullString{String: testURL, Valid: true},
                SourceName:    sql.NullString{String: "Test Source", Valid: true},
                PublishedDate: sql.NullTime{Time: time.Now(), Valid: true},
                Summary:       sql.NullString{String: "Test summary", Valid: true},
                Entities:      `{"entities": ["test"]}`,
                ContentType:   sql.NullString{String: "article", Valid: true},
                Topics:        `{"topics": ["tech"]}`,
                Status:        sql.NullString{String: "unread", Valid: true},
                StoryGroupID:  sql.NullString{String: "group-1", Valid: true},
        }

        created, err := queries.CreateArticle(ctx, params)
        require.NoError(t, err)

        found, err := queries.GetArticleByUrl(ctx, sql.NullString{String: testURL, Valid: true})
        require.NoError(t, err)

        assert.Equal(t, created.ID, found.ID)
        assert.Equal(t, created.Title.String, found.Title.String)
        assert.Equal(t, created.Url.String, found.Url.String)
}

func TestGetArticleByUrl_NotFound(t *testing.T) {
        _, queries := setupTestDB(t)
        ctx := context.Background()

        _, err := queries.GetArticleByUrl(ctx, sql.NullString{String: "https://nonexistent.com", Valid: true})
        assert.Error(t, err)
        assert.Equal(t, sql.ErrNoRows, err)
}

func TestListArticles(t *testing.T) {
        _, queries := setupTestDB(t)
        ctx := context.Background()

        articles, err := queries.ListArticles(ctx)
        require.NoError(t, err)
        assert.Empty(t, articles)

        params1 := CreateArticleParams{
                Title:         sql.NullString{String: "Article 1", Valid: true},
                Url:           sql.NullString{String: "https://example.com/1", Valid: true},
                SourceName:    sql.NullString{String: "Source 1", Valid: true},
                PublishedDate: sql.NullTime{Time: time.Now(), Valid: true},
                Summary:       sql.NullString{String: "Summary 1", Valid: true},
                Entities:      `{"entities": ["test1"]}`,
                ContentType:   sql.NullString{String: "article", Valid: true},
                Topics:        `{"topics": ["tech"]}`,
                Status:        sql.NullString{String: "unread", Valid: true},
                StoryGroupID:  sql.NullString{String: "group-1", Valid: true},
        }

        params2 := CreateArticleParams{
                Title:         sql.NullString{String: "Article 2", Valid: true},
                Url:           sql.NullString{String: "https://example.com/2", Valid: true},
                SourceName:    sql.NullString{String: "Source 2", Valid: true},
                PublishedDate: sql.NullTime{Time: time.Now().Add(time.Hour), Valid: true},
                Summary:       sql.NullString{String: "Summary 2", Valid: true},
                Entities:      `{"entities": ["test2"]}`,
                ContentType:   sql.NullString{String: "article", Valid: true},
                Topics:        `{"topics": ["science"]}`,
                Status:        sql.NullString{String: "read", Valid: true},
                StoryGroupID:  sql.NullString{String: "group-2", Valid: true},
        }

        _, err = queries.CreateArticle(ctx, params1)
        require.NoError(t, err)

        _, err = queries.CreateArticle(ctx, params2)
        require.NoError(t, err)

        articles, err = queries.ListArticles(ctx)
        require.NoError(t, err)
        assert.Len(t, articles, 2)

        assert.Equal(t, "Article 1", articles[0].Title.String)
        assert.Equal(t, "Article 2", articles[1].Title.String)
}

func TestUniqueUrlConstraint(t *testing.T) {
        _, queries := setupTestDB(t)
        ctx := context.Background()

        duplicateURL := "https://example.com/duplicate"

        params := CreateArticleParams{
                Title:         sql.NullString{String: "First Article", Valid: true},
                Url:           sql.NullString{String: duplicateURL, Valid: true},
                SourceName:    sql.NullString{String: "Test Source", Valid: true},
                PublishedDate: sql.NullTime{Time: time.Now(), Valid: true},
                Summary:       sql.NullString{String: "Test summary", Valid: true},
                Entities:      `{"entities": ["test"]}`,
                ContentType:   sql.NullString{String: "article", Valid: true},
                Topics:        `{"topics": ["tech"]}`,
                Status:        sql.NullString{String: "unread", Valid: true},
                StoryGroupID:  sql.NullString{String: "group-1", Valid: true},
        }

        _, err := queries.CreateArticle(ctx, params)
        require.NoError(t, err)

        params.Title = sql.NullString{String: "Second Article", Valid: true}
        _, err = queries.CreateArticle(ctx, params)
        assert.Error(t, err)
}

func TestDefaultStatus(t *testing.T) {
        db, queries := setupTestDB(t)
        ctx := context.Background()

        _, err := db.Exec(`
                INSERT INTO articles (title, url, source_name, published_date, summary, entities, content_type, topics, story_group_id)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        `, "Test Article", "https://example.com/default-status", "Test Source", time.Now(), "Test summary", `{"entities": ["test"]}`, "article", `{"topics": ["tech"]}`, "group-1")
        require.NoError(t, err)

        article, err := queries.GetArticleByUrl(ctx, sql.NullString{String: "https://example.com/default-status", Valid: true})
        require.NoError(t, err)

        assert.True(t, article.Status.Valid)
        assert.Equal(t, "unread", article.Status.String)
}

func TestSchemaCreation(t *testing.T) {
        db, err := sql.Open("sqlite", ":memory:")
        require.NoError(t, err)
        defer db.Close()

        err = InitSchema(db)
        require.NoError(t, err)

        var tableName string
        err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='articles'").Scan(&tableName)
        require.NoError(t, err)
        assert.Equal(t, "articles", tableName)

        rows, err := db.Query("PRAGMA table_info(articles)")
        require.NoError(t, err)
        defer rows.Close()

        expectedColumns := map[string]bool{
                "id":              false,
                "title":           false,
                "url":             false,
                "source_name":     false,
                "published_date":  false,
                "summary":         false,
                "entities":        false,
                "content_type":    false,
                "topics":          false,
                "status":          false,
                "story_group_id":  false,
        }

        for rows.Next() {
                var cid int
                var name, dataType string
                var notNull int
                var defaultValue sql.NullString
                var pk int

                err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
                require.NoError(t, err)

                if _, exists := expectedColumns[name]; exists {
                        expectedColumns[name] = true
                }

                if name == "status" && defaultValue.Valid {
                        assert.Equal(t, "'unread'", defaultValue.String)
                }
        }

        for column, found := range expectedColumns {
                assert.True(t, found, "Column %s not found in table", column)
        }
}

func TestDatabaseHelpers(t *testing.T) {
        db, queries, err := Open(":memory:")
        require.NoError(t, err)
        defer db.Close()

        assert.NotNil(t, db)
        assert.NotNil(t, queries)

        err = db.Ping()
        require.NoError(t, err)

        err = InitSchema(db)
        require.NoError(t, err)
}
