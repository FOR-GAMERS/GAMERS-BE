ALTER TABLE users
ADD COLUMN riot_name VARCHAR(32) NULL,
ADD COLUMN riot_tag VARCHAR(8) NULL,
ADD COLUMN region VARCHAR(10) NULL,
ADD COLUMN current_tier INT NULL,
ADD COLUMN current_tier_patched VARCHAR(32) NULL,
ADD COLUMN elo INT NULL,
ADD COLUMN ranking_in_tier INT NULL,
ADD COLUMN peak_tier INT NULL,
ADD COLUMN peak_tier_patched VARCHAR(32) NULL,
ADD COLUMN valorant_updated_at TIMESTAMP NULL;

CREATE INDEX idx_users_riot_name_tag ON users(riot_name, riot_tag);
