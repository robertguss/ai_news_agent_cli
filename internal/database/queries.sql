-- name: CreateArticle :one
INSERT INTO articles (
    title,
    url,
    source_name,
    published_date,
    summary,
    entities,
    content_type,
    topics,
    status,
    story_group_id
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetArticleByUrl :one
SELECT * FROM articles WHERE url = ? LIMIT 1;

-- name: ListArticles :many
SELECT * FROM articles;

-- name: ListUnreadArticles :many
SELECT * FROM articles WHERE status != 'read' ORDER BY published_date DESC;

-- name: ListAllArticles :many
SELECT * FROM articles ORDER BY published_date DESC;

-- name: ListArticlesBySource :many
SELECT * FROM articles WHERE status != 'read' AND source_name = ? ORDER BY published_date DESC;

-- name: ListArticlesByTopic :many
SELECT * FROM articles WHERE status != 'read' AND JSON_EXTRACT(topics, '$') LIKE '%' || ? || '%' ORDER BY published_date DESC;

-- name: ListArticlesBySourceAndTopic :many
SELECT * FROM articles WHERE status != 'read' AND source_name = ? AND JSON_EXTRACT(topics, '$') LIKE '%' || ? || '%' ORDER BY published_date DESC;

-- name: ListAllArticlesBySource :many
SELECT * FROM articles WHERE source_name = ? ORDER BY published_date DESC;

-- name: ListAllArticlesByTopic :many
SELECT * FROM articles WHERE JSON_EXTRACT(topics, '$') LIKE '%' || ? || '%' ORDER BY published_date DESC;

-- name: ListAllArticlesBySourceAndTopic :many
SELECT * FROM articles WHERE source_name = ? AND JSON_EXTRACT(topics, '$') LIKE '%' || ? || '%' ORDER BY published_date DESC;

-- name: MarkArticlesAsRead :exec
UPDATE articles SET status = 'read' WHERE id IN (sqlc.slice('ids'));

-- name: MarkArticleAsRead :exec
UPDATE articles SET status = 'read' WHERE id = ?;

-- name: GetArticle :one
SELECT * FROM articles WHERE id = ? LIMIT 1;

-- name: UpdateArticleStatus :exec
UPDATE articles SET status = ? WHERE id = ?;
