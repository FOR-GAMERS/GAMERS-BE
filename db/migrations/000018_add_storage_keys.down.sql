-- Remove banner_key from contests table
ALTER TABLE contests DROP COLUMN banner_key;

-- Remove profile_key from users table
ALTER TABLE users DROP COLUMN profile_key;
