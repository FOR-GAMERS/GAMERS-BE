-- Drop tables first (in reverse dependency order)
DROP TABLE IF EXISTS match_player_stats;
DROP TABLE IF EXISTS match_results;

-- Drop indexes conditionally
SET @idx_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'games' AND INDEX_NAME = 'idx_games_detection');
SET @sql = IF(@idx_exists > 0, 'DROP INDEX idx_games_detection ON games', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'games' AND INDEX_NAME = 'idx_games_scheduled');
SET @sql = IF(@idx_exists > 0, 'DROP INDEX idx_games_scheduled ON games', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Drop columns conditionally
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'games' AND COLUMN_NAME = 'scheduled_start_time');
SET @sql = IF(@col_exists > 0, 'ALTER TABLE games DROP COLUMN scheduled_start_time', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'games' AND COLUMN_NAME = 'detection_window_minutes');
SET @sql = IF(@col_exists > 0, 'ALTER TABLE games DROP COLUMN detection_window_minutes', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'games' AND COLUMN_NAME = 'detected_match_id');
SET @sql = IF(@col_exists > 0, 'ALTER TABLE games DROP COLUMN detected_match_id', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'games' AND COLUMN_NAME = 'detection_status');
SET @sql = IF(@col_exists > 0, 'ALTER TABLE games DROP COLUMN detection_status', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
