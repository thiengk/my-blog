-- Drop indexes first, then table
DROP INDEX IF EXISTS idx_subscribers_token;
DROP INDEX IF EXISTS idx_subscribers_status;
DROP INDEX IF EXISTS idx_subscribers_email;
DROP TABLE IF EXISTS newsletter_subscribers;
