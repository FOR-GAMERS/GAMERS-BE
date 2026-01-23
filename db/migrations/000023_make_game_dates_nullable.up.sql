-- Make started_at and ended_at nullable for tournament games
-- Tournament games are created before the schedule is determined

-- Drop the check constraint first
ALTER TABLE games DROP CONSTRAINT chk_game_dates;

-- Modify columns to allow NULL
ALTER TABLE games MODIFY COLUMN started_at DATETIME NULL;
ALTER TABLE games MODIFY COLUMN ended_at DATETIME NULL;

-- Add updated check constraint that allows NULL values
ALTER TABLE games ADD CONSTRAINT chk_game_dates
    CHECK (started_at IS NULL OR ended_at IS NULL OR started_at < ended_at);
