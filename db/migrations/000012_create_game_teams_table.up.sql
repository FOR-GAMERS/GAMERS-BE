-- Create game_teams table (game participation)
-- This table stores which teams are participating in which games (for ranking/grading)
CREATE TABLE game_teams (
    game_team_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    game_id BIGINT NOT NULL,
    team_id BIGINT NOT NULL,
    grade INT NULL,

    -- Foreign key constraints
    CONSTRAINT fk_game_teams_game
        FOREIGN KEY (game_id) REFERENCES games(game_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_game_teams_team
        FOREIGN KEY (team_id) REFERENCES teams(team_id)
        ON DELETE CASCADE,

    -- Unique constraint to prevent duplicate team entries in a game
    CONSTRAINT uq_game_teams_game_team UNIQUE (game_id, team_id),

    -- Unique constraint to prevent duplicate grades in the same game
    CONSTRAINT uq_game_teams_game_grade UNIQUE (game_id, grade),

    -- Check constraint for grade (must be at least 1 if not null)
    CONSTRAINT chk_game_teams_grade
        CHECK (grade IS NULL OR grade >= 1),

    -- Indexes
    INDEX idx_game_teams_game_id (game_id),
    INDEX idx_game_teams_team_id (team_id),
    INDEX idx_game_teams_grade (grade)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
