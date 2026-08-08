ALTER TABLE users
    ADD COLUMN IF NOT EXISTS weight_kg DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS height_cm DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS age INTEGER,
    ADD COLUMN IF NOT EXISTS sex TEXT,
    ADD COLUMN IF NOT EXISTS activity_level TEXT;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_sex_check;
ALTER TABLE users
    ADD CONSTRAINT users_sex_check
    CHECK (sex IS NULL OR sex IN ('male', 'female'));

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_activity_level_check;
ALTER TABLE users
    ADD CONSTRAINT users_activity_level_check
    CHECK (
        activity_level IS NULL OR activity_level IN (
            'sedentary',
            'light',
            'moderate',
            'active',
            'very_active'
        )
    );

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_weight_kg_check;
ALTER TABLE users
    ADD CONSTRAINT users_weight_kg_check
    CHECK (weight_kg IS NULL OR (weight_kg > 0 AND weight_kg < 500));

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_height_cm_check;
ALTER TABLE users
    ADD CONSTRAINT users_height_cm_check
    CHECK (height_cm IS NULL OR (height_cm > 0 AND height_cm < 300));

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_age_check;
ALTER TABLE users
    ADD CONSTRAINT users_age_check
    CHECK (age IS NULL OR (age >= 10 AND age <= 120));
