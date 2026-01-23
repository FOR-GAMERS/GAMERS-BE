-- Revert started_at and ended_at to NOT NULL

-- Drop the nullable check constraint
ALTER TABLE games DROP CONSTRAINT IF EXISTS chk_game_dates;

-- Update any NULL values to a default timestamp before making NOT NULL
UPDATE games SET started_at = NOW() WHERE started_at IS NULL;
UPDATE games SET ended_at = DATE_ADD(NOW(), INTERVAL 1 HOUR) WHERE ended_at IS NULL;

-- Modify columns back to NOT NULL
ALTER TABLE games MODIFY COLUMN started_at DATETIME NOT NULL;
ALTER TABLE games MODIFY COLUMN ended_at DATETIME NOT NULL;

-- Add original check constraint
ALTER TABLE games ADD CONSTRAINT chk_game_dates
    CHECK (started_at < ended_at);
