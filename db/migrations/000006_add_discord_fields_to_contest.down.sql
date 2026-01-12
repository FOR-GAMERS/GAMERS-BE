DROP INDEX idx_contest_discord_guild ON contests;

ALTER TABLE contests
DROP COLUMN discord_text_channel_id,
DROP COLUMN discord_guild_id;
