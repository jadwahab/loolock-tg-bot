CREATE TABLE IF NOT EXISTS leaderboard (
    id SERIAL PRIMARY KEY,
    amount_locked BIGINT NOT NULL,
    paymail VARCHAR(255) UNIQUE NOT NULL,
    public_key CHAR(66) NOT NULL,
    telegram_username VARCHAR(255),
    is_verified BOOLEAN DEFAULT FALSE,
    challenge TEXT,
    signature TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Creating an index on paymail for efficient lookups.
CREATE INDEX IF NOT EXISTS idx_paymail ON leaderboard(paymail);
