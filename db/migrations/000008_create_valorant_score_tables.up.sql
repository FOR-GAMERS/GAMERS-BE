-- Create valorant_score_tables table
CREATE TABLE valorant_score_tables (
    score_table_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    radiant INT NOT NULL,
    immortal INT NOT NULL,
    ascendant INT NOT NULL,
    diamond INT NOT NULL,
    platinum INT NOT NULL,
    gold INT NOT NULL,
    silver INT NOT NULL,
    bronze INT NOT NULL,
    iron INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
