-- Create meal_members table
CREATE TABLE meal_members (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create meal_participations table
CREATE TABLE meal_participations (
    id BIGSERIAL PRIMARY KEY,
    member_id BIGINT NOT NULL REFERENCES meal_members(id) ON DELETE CASCADE,
    meal_type VARCHAR(20) NOT NULL CHECK (meal_type IN ('breakfast', 'lunch')),
    is_participating BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(member_id, meal_type)
);

-- Create meal_payments table
CREATE TABLE meal_payments (
    id BIGSERIAL PRIMARY KEY,
    member_id BIGINT NOT NULL REFERENCES meal_members(id) ON DELETE CASCADE,
    meal_type VARCHAR(20) NOT NULL CHECK (meal_type IN ('breakfast', 'lunch')),
    payment_date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for query performance
CREATE INDEX idx_meal_participations_member_id ON meal_participations(member_id);
CREATE INDEX idx_meal_participations_meal_type ON meal_participations(meal_type);
CREATE INDEX idx_meal_payments_member_id ON meal_payments(member_id);
CREATE INDEX idx_meal_payments_meal_type ON meal_payments(meal_type);
CREATE INDEX idx_meal_payments_date ON meal_payments(payment_date DESC);
CREATE INDEX idx_meal_members_active ON meal_members(is_active) WHERE deleted_at IS NULL;
