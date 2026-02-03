-- Add match detection fields to games table
ALTER TABLE games
    ADD COLUMN scheduled_start_time DATETIME NULL,
    ADD COLUMN detection_window_minutes INT NOT NULL DEFAULT 120,
    ADD COLUMN detected_match_id VARCHAR(255) NULL,
    ADD COLUMN detection_status VARCHAR(20) NOT NULL DEFAULT 'NONE';

CREATE INDEX idx_games_detection ON games(game_status, detection_status);
CREATE INDEX idx_games_scheduled ON games(scheduled_start_time, game_status);

-- Match results table
CREATE TABLE match_results (
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
CREATE TABLE match_player_stats (
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
