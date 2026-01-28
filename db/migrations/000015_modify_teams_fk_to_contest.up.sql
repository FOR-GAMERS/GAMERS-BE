-- Modify teams table to reference contest_id instead of game_id
-- This allows teams to be registered for a contest rather than a specific game

-- First, drop the existing foreign key constraint
ALTER TABLE teams DROP FOREIGN KEY fk_teams_game;

-- Drop the existing unique constraint
ALTER TABLE teams DROP INDEX uq_teams_game_name;

-- Drop the index on game_id
ALTER TABLE teams DROP INDEX idx_teams_game_id;

-- Rename game_id column to contest_id
ALTER TABLE teams CHANGE COLUMN game_id contest_id BIGINT NOT NULL;

-- Add new foreign key constraint to contests table
ALTER TABLE teams ADD CONSTRAINT fk_teams_contest
    FOREIGN KEY (contest_id) REFERENCES contests(contest_id)
    ON DELETE CASCADE;

-- Add new unique constraint (team name unique within a contest)
ALTER TABLE teams ADD CONSTRAINT uq_teams_contest_name UNIQUE (contest_id, team_name);

-- Add new index on contest_id
ALTER TABLE teams ADD INDEX idx_teams_contest_id (contest_id);
