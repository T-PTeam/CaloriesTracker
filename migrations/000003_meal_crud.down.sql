ALTER TABLE meal_items
    DROP CONSTRAINT IF EXISTS meal_items_category_check;

ALTER TABLE meal_items
    DROP COLUMN IF EXISTS category;

ALTER TABLE meals
    DROP COLUMN IF EXISTS sort_order;
