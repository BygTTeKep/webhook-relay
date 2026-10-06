CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY,
    event_id TEXT UNIQUE NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP
);