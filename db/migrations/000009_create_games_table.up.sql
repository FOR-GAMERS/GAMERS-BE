-- Create games table
CREATE TABLE games (
    game_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    contest_id BIGINT NOT NULL,
    game_status VARCHAR(16) NOT NULL,
    game_team_type VARCHAR(16) NOT NULL,
    started_at DATETIME NOT NULL,
    ended_at DATETIME NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    -- Foreign key constraint with CASCADE delete
    CONSTRAINT fk_games_contest
        FOREIGN KEY (contest_id) REFERENCES contests(contest_id)
        ON DELETE CASCADE,

    -- Check constraints
    CONSTRAINT chk_game_status
        CHECK (game_status IN ('PENDING', 'ACTIVE', 'FINISHED', 'CANCELLED')),
    CONSTRAINT chk_game_team_type
        CHECK (game_team_type IN ('SINGLE', 'DUO', 'TRIO', 'FULL', 'HURUPA')),
    CONSTRAINT chk_game_dates
        CHECK (started_at < ended_at),

    -- Indexes
    INDEX idx_game_contest_id (contest_id),
    INDEX idx_game_status (game_status),
    INDEX idx_game_dates (started_at, ended_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
