-- Create teams table (team entity)
-- This table stores the team itself, not the members
CREATE TABLE teams (
    team_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    game_id BIGINT NOT NULL,
    team_name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    -- Foreign key constraint
    CONSTRAINT fk_teams_game
        FOREIGN KEY (game_id) REFERENCES games(game_id)
        ON DELETE CASCADE,

    -- Unique constraint to prevent duplicate team names in the same game
    CONSTRAINT uq_teams_game_name UNIQUE (game_id, team_name),

    -- Indexes
    INDEX idx_teams_game_id (game_id),
    INDEX idx_teams_team_name (team_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
