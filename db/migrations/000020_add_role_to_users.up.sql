-- Add role column to users table
ALTER TABLE users ADD COLUMN role VARCHAR(16) NOT NULL DEFAULT 'USER';

-- Create index for role column for efficient filtering
CREATE INDEX idx_users_role ON users(role);
