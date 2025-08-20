-- Migration to add analysis_status column and fix corrupted data
-- This migration adds the analysis_status column and repairs articles
-- that had their read/unread status corrupted by AI analysis status

-- Add the new analysis_status column
ALTER TABLE articles ADD COLUMN analysis_status TEXT DEFAULT 'unprocessed';

-- Back-fill and repair previously corrupted rows
-- Articles with status 'pending', 'completed', or 'unprocessed' should have:
-- - analysis_status set to their current status value
-- - status reset to 'unread' (since they were never actually read)
UPDATE articles 
SET analysis_status = status,
    status = 'unread'
WHERE status IN ('unprocessed', 'pending', 'completed');

-- Verify the migration worked
-- This should show 0 rows if successful
SELECT COUNT(*) as corrupted_rows 
FROM articles 
WHERE status IN ('unprocessed', 'pending', 'completed');
