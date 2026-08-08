CREATE TABLE IF NOT EXISTS users (
    id          BIGSERIAL PRIMARY KEY,
    telegram_id BIGINT NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS meals (
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    total_calories DOUBLE PRECISION NOT NULL DEFAULT 0,
    total_protein  DOUBLE PRECISION NOT NULL DEFAULT 0,
    total_fat      DOUBLE PRECISION NOT NULL DEFAULT 0,
    total_carbs    DOUBLE PRECISION NOT NULL DEFAULT 0,
    raw_text       TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_meals_user_created
    ON meals (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS meal_items (
    id         BIGSERIAL PRIMARY KEY,
    meal_id    BIGINT NOT NULL REFERENCES meals(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    weight_g   DOUBLE PRECISION NOT NULL DEFAULT 0,
    calories   DOUBLE PRECISION NOT NULL DEFAULT 0,
    protein    DOUBLE PRECISION NOT NULL DEFAULT 0,
    fat        DOUBLE PRECISION NOT NULL DEFAULT 0,
    carbs      DOUBLE PRECISION NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_meal_items_meal_id
    ON meal_items (meal_id);
