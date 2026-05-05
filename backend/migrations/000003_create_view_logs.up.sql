-- Create view_logs table to track views for duplicate detection
CREATE TABLE view_logs (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(255) NOT NULL,
    ip_hash VARCHAR(64) NOT NULL,
    viewed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Composite index on slug + ip_hash for duplicate view detection
CREATE INDEX idx_view_logs_slug_ip ON view_logs(slug, ip_hash);

-- Index on viewed_at for cleanup of old logs (older than 24h)
CREATE INDEX idx_view_logs_viewed_at ON view_logs(viewed_at);
