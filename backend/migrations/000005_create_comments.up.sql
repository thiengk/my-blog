-- Create comments table to store blog post comments
CREATE TABLE comments (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(255) NOT NULL,
    author_name VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    ip_hash VARCHAR(64) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index on slug for fetching comments by post
CREATE INDEX idx_comments_slug ON comments(slug);

-- Composite index on slug + created_at for chronological listing per post
CREATE INDEX idx_comments_slug_created ON comments(slug, created_at ASC);
