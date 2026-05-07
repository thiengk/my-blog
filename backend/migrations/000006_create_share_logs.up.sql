-- Create share_logs table to track shares with platform information
CREATE TABLE share_logs (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(255) NOT NULL,
    platform VARCHAR(20) NOT NULL,
    ip_hash VARCHAR(64) NOT NULL,
    shared_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Composite index on slug + platform for platform-specific share queries
CREATE INDEX idx_share_logs_slug_platform ON share_logs(slug, platform);
