DROP TABLE IF EXISTS match_player_stats;
DROP TABLE IF EXISTS match_results;

DROP INDEX idx_games_detection ON games;
DROP INDEX idx_games_scheduled ON games;

ALTER TABLE games
    DROP COLUMN scheduled_start_time,
    DROP COLUMN detection_window_minutes,
    DROP COLUMN detected_match_id,
    DROP COLUMN detection_status;
