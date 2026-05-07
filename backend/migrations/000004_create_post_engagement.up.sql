-- Create post_engagement table to store engagement counts for each blog post
CREATE TABLE post_engagement (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(255) NOT NULL UNIQUE,
    like_count BIGINT NOT NULL DEFAULT 0,
    comment_count BIGINT NOT NULL DEFAULT 0,
    share_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index on slug for fast lookup by post slug
CREATE INDEX idx_post_engagement_slug ON post_engagement(slug);

-- Computed index on engagement score for recommendation queries
-- Score formula: likes * 1 + comments * 2 + shares * 3
CREATE INDEX idx_post_engagement_score ON post_engagement(
    (like_count * 1 + comment_count * 2 + share_count * 3) DESC
);
