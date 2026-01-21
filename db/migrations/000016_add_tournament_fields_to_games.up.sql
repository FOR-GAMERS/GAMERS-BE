-- Add tournament bracket fields to games table
ALTER TABLE games
    ADD COLUMN round INT NULL COMMENT 'Tournament round number (1 = first round, 2 = quarter finals, etc.)',
    ADD COLUMN match_number INT NULL COMMENT 'Match number within the round',
    ADD COLUMN next_game_id BIGINT NULL COMMENT 'The game that the winner advances to',
    ADD COLUMN bracket_position INT NULL COMMENT 'Position in the bracket (for display purposes)';

-- Add foreign key for next_game_id
ALTER TABLE games ADD CONSTRAINT fk_games_next_game
    FOREIGN KEY (next_game_id) REFERENCES games(game_id)
    ON DELETE SET NULL;

-- Add indexes for tournament queries
ALTER TABLE games ADD INDEX idx_games_round (round);
ALTER TABLE games ADD INDEX idx_games_match_number (match_number);
ALTER TABLE games ADD INDEX idx_games_next_game_id (next_game_id);
