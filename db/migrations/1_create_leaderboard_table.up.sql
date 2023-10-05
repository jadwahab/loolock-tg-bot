CREATE TABLE leaderboard (
    id SERIAL PRIMARY KEY,
    amount_locked BIGINT NOT NULL,
    paymail VARCHAR(255) UNIQUE NOT NULL,
    public_key CHAR(66) NOT NULL,
    telegram_username VARCHAR(255)
    is_verified BOOLEAN DEFAULT FALSE,
    challenge TEXT,
    signature TEXT,
    created_at TIMESTAMPTZ CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ CURRENT_TIMESTAMP
);

-- Creating an index on paymail for efficient lookups.
CREATE INDEX idx_paymail ON leaderboard(paymail);


CREATE TABLE leaderboard (
    id SERIAL PRIMARY KEY,
    paymail VARCHAR(255) UNIQUE NOT NULL, -- ensuring paymails are unique
    public_key CHAR(66) NOT NULL,
    amount_locked NUMERIC NOT NULL,
    last_updated TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    is_verified BOOLEAN DEFAULT FALSE, -- tracking if a user has been verified
    telegram_username VARCHAR(255) -- storing the Telegram username
);
CREATE INDEX idx_paymail ON leaderboard(paymail); -- indexing for faster lookups by paymail
