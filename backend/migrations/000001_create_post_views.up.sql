-- Create post_views table to store view count for each blog post
CREATE TABLE post_views (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(255) NOT NULL UNIQUE,
    view_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index on slug for fast lookup by post slug
CREATE INDEX idx_post_views_slug ON post_views(slug);

-- Index on view_count descending for popular posts queries
CREATE INDEX idx_post_views_count ON post_views(view_count DESC);
