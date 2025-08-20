-- Add content column to store full article content
-- This allows viewing articles in the TUI without real-time fetching

ALTER TABLE articles ADD COLUMN content TEXT;