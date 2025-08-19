CREATE TABLE IF NOT EXISTS articles (
    id INTEGER PRIMARY KEY,
    title TEXT,
    url TEXT UNIQUE,
    source_name TEXT,
    published_date DATETIME,
    summary TEXT,
    entities JSON,
    content_type TEXT,
    topics JSON,
    status TEXT DEFAULT 'unread',
    story_group_id TEXT
);
