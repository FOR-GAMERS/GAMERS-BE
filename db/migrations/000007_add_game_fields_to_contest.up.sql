-- Add game-related fields to contests table
ALTER TABLE contests
    ADD COLUMN game_type VARCHAR(32) NULL AFTER auto_start,
    ADD COLUMN game_point_table_id BIGINT NULL AFTER game_type,
    ADD COLUMN total_team_member INT NOT NULL DEFAULT 5 AFTER game_point_table_id;

-- Modify description column to TEXT for MDX content
ALTER TABLE contests
    MODIFY COLUMN description TEXT;

-- Add index for game_type
CREATE INDEX idx_contest_game_type ON contests(game_type);

-- Add constraint for game_type
ALTER TABLE contests
    ADD CONSTRAINT chk_game_type
    CHECK (game_type IS NULL OR game_type IN ('VALORANT', 'LOL'));
