-- Add the 'left_at' column
ALTER TABLE group_chat_users
ADD COLUMN left_at TIMESTAMPTZ DEFAULT NULL;

-- Drop the duplicate index if it exists
DROP INDEX IF EXISTS idx_user_id;

-- Create the correct index for username
CREATE INDEX IF NOT EXISTS idx_username ON group_chat_users(username);
