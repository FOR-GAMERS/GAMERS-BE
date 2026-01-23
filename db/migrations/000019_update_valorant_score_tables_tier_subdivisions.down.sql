-- Rollback: Revert tier subdivisions back to single tier columns

-- Add back the original single tier columns
ALTER TABLE valorant_score_tables
    ADD COLUMN immortal INT NOT NULL DEFAULT 0 AFTER radiant,
    ADD COLUMN ascendant INT NOT NULL DEFAULT 0 AFTER immortal,
    ADD COLUMN diamond INT NOT NULL DEFAULT 0 AFTER ascendant,
    ADD COLUMN platinum INT NOT NULL DEFAULT 0 AFTER diamond,
    ADD COLUMN gold INT NOT NULL DEFAULT 0 AFTER platinum,
    ADD COLUMN silver INT NOT NULL DEFAULT 0 AFTER gold,
    ADD COLUMN bronze INT NOT NULL DEFAULT 0 AFTER silver,
    ADD COLUMN iron INT NOT NULL DEFAULT 0 AFTER bronze;

-- Migrate data: use tier 1 values for the single tier columns
UPDATE valorant_score_tables SET
    immortal = immortal_1,
    ascendant = ascendant_1,
    diamond = diamond_1,
    platinum = platinum_1,
    gold = gold_1,
    silver = silver_1,
    bronze = bronze_1,
    iron = iron_1;

-- Drop the subdivision columns
ALTER TABLE valorant_score_tables
    DROP COLUMN immortal_3,
    DROP COLUMN immortal_2,
    DROP COLUMN immortal_1,
    DROP COLUMN ascendant_3,
    DROP COLUMN ascendant_2,
    DROP COLUMN ascendant_1,
    DROP COLUMN diamond_3,
    DROP COLUMN diamond_2,
    DROP COLUMN diamond_1,
    DROP COLUMN platinum_3,
    DROP COLUMN platinum_2,
    DROP COLUMN platinum_1,
    DROP COLUMN gold_3,
    DROP COLUMN gold_2,
    DROP COLUMN gold_1,
    DROP COLUMN silver_3,
    DROP COLUMN silver_2,
    DROP COLUMN silver_1,
    DROP COLUMN bronze_3,
    DROP COLUMN bronze_2,
    DROP COLUMN bronze_1,
    DROP COLUMN iron_3,
    DROP COLUMN iron_2,
    DROP COLUMN iron_1;
