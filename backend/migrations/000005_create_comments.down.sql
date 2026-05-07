-- Drop indexes first, then table
DROP INDEX IF EXISTS idx_comments_slug_created;
DROP INDEX IF EXISTS idx_comments_slug;
DROP TABLE IF EXISTS comments;
