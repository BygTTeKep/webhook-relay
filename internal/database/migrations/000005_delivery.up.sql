CREATE TABLE IF NOT EXISTS delivery (
    id BIGSERIAL PRIMARY KEY,
    event_id UUID REFERENCES events(id) ON DELETE CASCADE,
    subscription_id UUID REFERENCES subscription(id) ON DELETE CASCADE,
    status VARCHAR(30) NOT NULL,
    attempts INT NOT NULL,
    last_attempt_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);