CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY,
    url text NOT NULL,
    secret text NOT NULL,
    active boolean NOT NULL DEFAULT TRUE,
    created_at timestamp CURRENT_TIMESTAMP
);