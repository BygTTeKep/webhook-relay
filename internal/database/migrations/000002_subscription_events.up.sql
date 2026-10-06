CREATE TABLE IF NOT EXISTS subscription_events (
    id BIGSERIAL PRIMARY KEY,
    subscription_id UUID REFERENCES subscription(id) ON DELETE CASCADE
    event_type VARCHAR(255) NOT NULL
);