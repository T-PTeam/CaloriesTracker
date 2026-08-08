ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_sex_check;
ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_activity_level_check;
ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_weight_kg_check;
ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_height_cm_check;
ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_age_check;

ALTER TABLE users
    DROP COLUMN IF EXISTS weight_kg,
    DROP COLUMN IF EXISTS height_cm,
    DROP COLUMN IF EXISTS age,
    DROP COLUMN IF EXISTS sex,
    DROP COLUMN IF EXISTS activity_level;
