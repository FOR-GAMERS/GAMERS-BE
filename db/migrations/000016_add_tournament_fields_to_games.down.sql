-- Remove tournament bracket fields from games table
ALTER TABLE games DROP FOREIGN KEY fk_games_next_game;

ALTER TABLE games DROP INDEX idx_games_round;
ALTER TABLE games DROP INDEX idx_games_match_number;
ALTER TABLE games DROP INDEX idx_games_next_game_id;

ALTER TABLE games
    DROP COLUMN round,
    DROP COLUMN match_number,
    DROP COLUMN next_game_id,
    DROP COLUMN bracket_position;
