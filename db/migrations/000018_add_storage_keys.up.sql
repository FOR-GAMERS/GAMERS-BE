-- Add banner_key to contests table for R2 storage reference
ALTER TABLE contests ADD COLUMN banner_key VARCHAR(512) NULL AFTER thumbnail;

-- Add profile_key to users table for R2 storage reference
ALTER TABLE users ADD COLUMN profile_key VARCHAR(512) NULL AFTER avatar;
