-- Drop indexes first, then table
DROP INDEX IF EXISTS idx_post_views_count;
DROP INDEX IF EXISTS idx_post_views_slug;
DROP TABLE IF EXISTS post_views;
