CREATE TABLE IF NOT EXISTS group_chat_users (
    id SERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL,  -- Telegram group chat ID
    user_id BIGINT NOT NULL,  -- Telegram user ID
    username VARCHAR(255),    -- Optional: Telegram username for reference
    joined_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- For fast lookups, especially if this table grows large
CREATE INDEX IF NOT EXISTS idx_chat_id ON group_chat_users(chat_id);
CREATE INDEX IF NOT EXISTS idx_user_id ON group_chat_users(user_id);
CREATE INDEX IF NOT EXISTS idx_user_id ON group_chat_users(username);
