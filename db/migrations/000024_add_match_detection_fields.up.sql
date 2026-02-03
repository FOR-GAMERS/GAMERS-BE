-- Add match detection fields to games table
-- Note: Using conditional approach to handle partial migrations

-- Add scheduled_start_time if not exists
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'games' AND COLUMN_NAME = 'scheduled_start_time');
SET @sql = IF(@col_exists = 0, 'ALTER TABLE games ADD COLUMN scheduled_start_time DATETIME NULL', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add detection_window_minutes if not exists
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'games' AND COLUMN_NAME = 'detection_window_minutes');
SET @sql = IF(@col_exists = 0, 'ALTER TABLE games ADD COLUMN detection_window_minutes INT NOT NULL DEFAULT 120', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add detected_match_id if not exists
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'games' AND COLUMN_NAME = 'detected_match_id');
SET @sql = IF(@col_exists = 0, 'ALTER TABLE games ADD COLUMN detected_match_id VARCHAR(255) NULL', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add detection_status if not exists
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'games' AND COLUMN_NAME = 'detection_status');
SET @sql = IF(@col_exists = 0, "ALTER TABLE games ADD COLUMN detection_status VARCHAR(20) NOT NULL DEFAULT 'NONE'", 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add idx_games_detection if not exists
SET @idx_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'games' AND INDEX_NAME = 'idx_games_detection');
SET @sql = IF(@idx_exists = 0, 'CREATE INDEX idx_games_detection ON games(game_status, detection_status)', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add idx_games_scheduled if not exists
SET @idx_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'games' AND INDEX_NAME = 'idx_games_scheduled');
SET @sql = IF(@idx_exists = 0, 'CREATE INDEX idx_games_scheduled ON games(scheduled_start_time, game_status)', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Match results table
CREATE TABLE IF NOT EXISTS match_results (
    match_result_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    game_id         BIGINT NOT NULL,
    valorant_match_id VARCHAR(255) NOT NULL,
    map_name        VARCHAR(50),
    rounds_played   INT NOT NULL,
    winner_team_id  BIGINT NOT NULL,
    loser_team_id   BIGINT NOT NULL,
    winner_score    INT NOT NULL,
    loser_score     INT NOT NULL,
    game_started_at DATETIME NOT NULL,
    game_duration   INT NOT NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE INDEX idx_match_results_game (game_id),
    INDEX idx_match_results_valorant (valorant_match_id),
    CONSTRAINT fk_match_results_game FOREIGN KEY (game_id) REFERENCES games(game_id),
    CONSTRAINT fk_match_results_winner FOREIGN KEY (winner_team_id) REFERENCES teams(team_id),
    CONSTRAINT fk_match_results_loser FOREIGN KEY (loser_team_id) REFERENCES teams(team_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Match player stats table
CREATE TABLE IF NOT EXISTS match_player_stats (
    match_player_stat_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    match_result_id      BIGINT NOT NULL,
    user_id              BIGINT NOT NULL,
    team_id              BIGINT NOT NULL,
    agent_name           VARCHAR(50),
    kills                INT NOT NULL DEFAULT 0,
    deaths               INT NOT NULL DEFAULT 0,
    assists              INT NOT NULL DEFAULT 0,
    score                INT NOT NULL DEFAULT 0,
    headshots            INT NOT NULL DEFAULT 0,
    bodyshots            INT NOT NULL DEFAULT 0,
    legshots             INT NOT NULL DEFAULT 0,

    INDEX idx_match_player_stats_result (match_result_id),
    INDEX idx_match_player_stats_user (user_id),
    CONSTRAINT fk_match_player_stats_result FOREIGN KEY (match_result_id) REFERENCES match_results(match_result_id),
    CONSTRAINT fk_match_player_stats_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_match_player_stats_team FOREIGN KEY (team_id) REFERENCES teams(team_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
