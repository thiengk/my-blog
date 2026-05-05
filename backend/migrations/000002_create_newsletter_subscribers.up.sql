-- Create newsletter_subscribers table to store subscriber information
CREATE TABLE newsletter_subscribers (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(320) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    verification_token VARCHAR(64),
    subscribed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    unsubscribed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index on email for fast lookup
CREATE INDEX idx_subscribers_email ON newsletter_subscribers(email);

-- Index on status for filtering active/pending/unsubscribed subscribers
CREATE INDEX idx_subscribers_status ON newsletter_subscribers(status);

-- Index on verification_token for email verification lookup
CREATE INDEX idx_subscribers_token ON newsletter_subscribers(verification_token);
