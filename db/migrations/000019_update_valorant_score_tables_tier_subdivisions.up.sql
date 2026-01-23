-- Update valorant_score_tables to support tier subdivisions (1, 2, 3) for each rank
-- Radiant remains single, all other tiers get 3 subdivisions

-- Add new columns for tier subdivisions
ALTER TABLE valorant_score_tables
    ADD COLUMN immortal_3 INT NOT NULL DEFAULT 0 AFTER radiant,
    ADD COLUMN immortal_2 INT NOT NULL DEFAULT 0 AFTER immortal_3,
    ADD COLUMN immortal_1 INT NOT NULL DEFAULT 0 AFTER immortal_2,
    ADD COLUMN ascendant_3 INT NOT NULL DEFAULT 0 AFTER immortal_1,
    ADD COLUMN ascendant_2 INT NOT NULL DEFAULT 0 AFTER ascendant_3,
    ADD COLUMN ascendant_1 INT NOT NULL DEFAULT 0 AFTER ascendant_2,
    ADD COLUMN diamond_3 INT NOT NULL DEFAULT 0 AFTER ascendant_1,
    ADD COLUMN diamond_2 INT NOT NULL DEFAULT 0 AFTER diamond_3,
    ADD COLUMN diamond_1 INT NOT NULL DEFAULT 0 AFTER diamond_2,
    ADD COLUMN platinum_3 INT NOT NULL DEFAULT 0 AFTER diamond_1,
    ADD COLUMN platinum_2 INT NOT NULL DEFAULT 0 AFTER platinum_3,
    ADD COLUMN platinum_1 INT NOT NULL DEFAULT 0 AFTER platinum_2,
    ADD COLUMN gold_3 INT NOT NULL DEFAULT 0 AFTER platinum_1,
    ADD COLUMN gold_2 INT NOT NULL DEFAULT 0 AFTER gold_3,
    ADD COLUMN gold_1 INT NOT NULL DEFAULT 0 AFTER gold_2,
    ADD COLUMN silver_3 INT NOT NULL DEFAULT 0 AFTER gold_1,
    ADD COLUMN silver_2 INT NOT NULL DEFAULT 0 AFTER silver_3,
    ADD COLUMN silver_1 INT NOT NULL DEFAULT 0 AFTER silver_2,
    ADD COLUMN bronze_3 INT NOT NULL DEFAULT 0 AFTER silver_1,
    ADD COLUMN bronze_2 INT NOT NULL DEFAULT 0 AFTER bronze_3,
    ADD COLUMN bronze_1 INT NOT NULL DEFAULT 0 AFTER bronze_2,
    ADD COLUMN iron_3 INT NOT NULL DEFAULT 0 AFTER bronze_1,
    ADD COLUMN iron_2 INT NOT NULL DEFAULT 0 AFTER iron_3,
    ADD COLUMN iron_1 INT NOT NULL DEFAULT 0 AFTER iron_2;

-- Migrate existing data: copy old single tier values to tier 1 (lowest in subdivision)
UPDATE valorant_score_tables SET
    immortal_3 = immortal,
    immortal_2 = immortal,
    immortal_1 = immortal,
    ascendant_3 = ascendant,
    ascendant_2 = ascendant,
    ascendant_1 = ascendant,
    diamond_3 = diamond,
    diamond_2 = diamond,
    diamond_1 = diamond,
    platinum_3 = platinum,
    platinum_2 = platinum,
    platinum_1 = platinum,
    gold_3 = gold,
    gold_2 = gold,
    gold_1 = gold,
    silver_3 = silver,
    silver_2 = silver,
    silver_1 = silver,
    bronze_3 = bronze,
    bronze_2 = bronze,
    bronze_1 = bronze,
    iron_3 = iron,
    iron_2 = iron,
    iron_1 = iron;

-- Drop old columns
ALTER TABLE valorant_score_tables
    DROP COLUMN immortal,
    DROP COLUMN ascendant,
    DROP COLUMN diamond,
    DROP COLUMN platinum,
    DROP COLUMN gold,
    DROP COLUMN silver,
    DROP COLUMN bronze,
    DROP COLUMN iron;
