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
