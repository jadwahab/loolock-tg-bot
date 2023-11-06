-- Remove the 'left_at' column
ALTER TABLE group_chat_users
DROP COLUMN IF EXISTS left_at;

-- Drop the corrected index for username
DROP INDEX IF EXISTS idx_username;

-- Re-create the dropped index for user_id if it was accidentally removed
CREATE INDEX IF NOT EXISTS idx_user_id ON group_chat_users(user_id);
