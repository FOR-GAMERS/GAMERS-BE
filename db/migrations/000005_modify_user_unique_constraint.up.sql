-- Remove the unique index on tag column
DROP INDEX IF EXISTS idx_users_tag ON users;

-- Create composite unique index on username and tag
CREATE UNIQUE INDEX idx_username_tag ON users(username, tag);
