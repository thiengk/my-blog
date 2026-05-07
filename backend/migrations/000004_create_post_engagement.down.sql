-- Drop indexes first, then table
DROP INDEX IF EXISTS idx_post_engagement_score;
DROP INDEX IF EXISTS idx_post_engagement_slug;
DROP TABLE IF EXISTS post_engagement;
