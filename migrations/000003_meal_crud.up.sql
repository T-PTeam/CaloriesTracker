ALTER TABLE meals
    ADD COLUMN IF NOT EXISTS sort_order INT NOT NULL DEFAULT 0;

ALTER TABLE meal_items
    ADD COLUMN IF NOT EXISTS category TEXT NOT NULL DEFAULT 'high_quality_protein';

ALTER TABLE meal_items
    DROP CONSTRAINT IF EXISTS meal_items_category_check;

ALTER TABLE meal_items
    ADD CONSTRAINT meal_items_category_check
    CHECK (category IN (
        'high_quality_protein',
        'long_acting_carbs',
        'lipids',
        'fast_acting_carbs'
    ));

WITH ranked AS (
    SELECT
        id,
        ROW_NUMBER() OVER (
            PARTITION BY user_id, (created_at AT TIME ZONE 'UTC')::date
            ORDER BY created_at ASC, id ASC
        ) - 1 AS rn
    FROM meals
)
UPDATE meals m
SET sort_order = ranked.rn
FROM ranked
WHERE m.id = ranked.id;
