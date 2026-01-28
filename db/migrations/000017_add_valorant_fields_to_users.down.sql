DROP INDEX idx_users_riot_name_tag ON users;

ALTER TABLE users
DROP COLUMN riot_name,
DROP COLUMN riot_tag,
DROP COLUMN region,
DROP COLUMN current_tier,
DROP COLUMN current_tier_patched,
DROP COLUMN elo,
DROP COLUMN ranking_in_tier,
DROP COLUMN peak_tier,
DROP COLUMN peak_tier_patched,
DROP COLUMN valorant_updated_at;
