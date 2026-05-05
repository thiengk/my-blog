-- Drop indexes first, then table
DROP INDEX IF EXISTS idx_view_logs_viewed_at;
DROP INDEX IF EXISTS idx_view_logs_slug_ip;
DROP TABLE IF EXISTS view_logs;
