-- Add chat_id column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS chat_id VARCHAR(100) DEFAULT '';

-- Add comment
COMMENT ON COLUMN users.chat_id IS 'Telegram Chat ID for bot notifications';
