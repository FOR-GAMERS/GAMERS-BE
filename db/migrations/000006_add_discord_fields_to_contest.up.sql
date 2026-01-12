ALTER TABLE contests
ADD COLUMN discord_guild_id VARCHAR(255) NULL AFTER auto_start,
ADD COLUMN discord_text_channel_id VARCHAR(255) NULL AFTER discord_guild_id;

CREATE INDEX idx_contest_discord_guild ON contests(discord_guild_id);
