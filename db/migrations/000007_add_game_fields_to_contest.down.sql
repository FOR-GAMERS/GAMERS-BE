-- Remove game-related fields from contests table
ALTER TABLE contests
    DROP CONSTRAINT IF EXISTS chk_game_type;

DROP INDEX idx_contest_game_type ON contests;

ALTER TABLE contests
    DROP COLUMN game_type,
    DROP COLUMN game_point_table_id,
    DROP COLUMN total_team_member;

-- Revert description column to VARCHAR(255)
ALTER TABLE contests
    MODIFY COLUMN description VARCHAR(255);
