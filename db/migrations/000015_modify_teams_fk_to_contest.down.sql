-- Rollback teams table to reference game_id instead of contest_id

-- Drop the foreign key constraint to contests
ALTER TABLE teams DROP FOREIGN KEY fk_teams_contest;

-- Drop the unique constraint
ALTER TABLE teams DROP INDEX uq_teams_contest_name;

-- Drop the index on contest_id
ALTER TABLE teams DROP INDEX idx_teams_contest_id;

-- Rename contest_id column back to game_id
ALTER TABLE teams CHANGE COLUMN contest_id game_id BIGINT NOT NULL;

-- Add back the foreign key constraint to games table
ALTER TABLE teams ADD CONSTRAINT fk_teams_game
    FOREIGN KEY (game_id) REFERENCES games(game_id)
    ON DELETE CASCADE;

-- Add back the unique constraint
ALTER TABLE teams ADD CONSTRAINT uq_teams_game_name UNIQUE (game_id, team_name);

-- Add back the index on game_id
ALTER TABLE teams ADD INDEX idx_teams_game_id (game_id);
