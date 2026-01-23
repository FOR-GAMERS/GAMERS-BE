-- Remove role column from users table
DROP INDEX idx_users_role ON users;
ALTER TABLE users DROP COLUMN role;
