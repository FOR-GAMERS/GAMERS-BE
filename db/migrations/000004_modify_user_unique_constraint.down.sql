-- Remove the composite unique index
DROP INDEX IF EXISTS idx_username_tag ON users;

-- Restore the unique index on tag column
CREATE UNIQUE INDEX idx_users_tag ON users(tag);
