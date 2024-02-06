CREATE TABLE IF NOT EXISTS top_50_chat ( -- Telegram group chat ID (-1001984238822)
    user_id BIGINT PRIMARY KEY,   -- Telegram user ID
    username VARCHAR(255),        -- Optional: Telegram username for reference
    tg_name VARCHAR(255),         -- Optional: Telegram name for reference
    member BOOLEAN DEFAULT FALSE, -- Whether the user is a member of the chat
    last_joined TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    last_left TIMESTAMPTZ DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_username ON top_50_chat(username);
