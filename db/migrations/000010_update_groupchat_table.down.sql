-- Remove the primary key constraint from user_id
ALTER TABLE group_chat_users DROP CONSTRAINT group_chat_users_pkey;

-- Add the 'id' column back
ALTER TABLE group_chat_users ADD COLUMN id SERIAL PRIMARY KEY;

-- Drop the 'tg_name' column if it exists
ALTER TABLE group_chat_users DROP COLUMN IF EXISTS tg_name;

-- Drop the 'member' column if it exists
ALTER TABLE group_chat_users DROP COLUMN IF EXISTS member;
