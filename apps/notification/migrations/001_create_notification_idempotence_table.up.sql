CREATE TABLE IF NOT EXISTS notification_idempotence (
    message_id   uuid PRIMARY KEY,
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
