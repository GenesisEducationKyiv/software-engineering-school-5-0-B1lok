CREATE TYPE status_enum AS ENUM ('pending', 'failed', 'success');
CREATE TYPE event_type_enum AS ENUM ('user_subscribed');

CREATE TABLE IF NOT EXISTS outbox (
    id SERIAL PRIMARY KEY,
    aggregate_id INTEGER NOT NULL,
    event_type event_type_enum NOT NULL,
    payload JSONB NOT NULL,
    status status_enum DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_outbox_status ON outbox(status);