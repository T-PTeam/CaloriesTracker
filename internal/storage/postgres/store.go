package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/root1/calories-tracker/internal/domain"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(ctx context.Context, databaseURL string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

const userSelectColumns = `
id, telegram_id, COALESCE(email, ''), COALESCE(password_hash, ''), created_at,
weight_kg, height_cm, age, COALESCE(sex, ''), COALESCE(activity_level, ''), COALESCE(language, '')`

func scanUser(row pgx.Row) (domain.User, error) {
	var u domain.User
	err := row.Scan(
		&u.ID,
		&u.TelegramID,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
		&u.WeightKg,
		&u.HeightCm,
		&u.Age,
		&u.Sex,
		&u.ActivityLevel,
		&u.Language,
	)
	return u, err
}

func (s *Store) EnsureUser(ctx context.Context, telegramID int64) (domain.User, error) {
	const q = `
INSERT INTO users (telegram_id)
VALUES ($1)
ON CONFLICT (telegram_id) DO UPDATE SET telegram_id = EXCLUDED.telegram_id
RETURNING ` + userSelectColumns

	u, err := scanUser(s.pool.QueryRow(ctx, q, telegramID))
	if err != nil {
		return domain.User{}, fmt.Errorf("ensure user: %w", err)
	}
	return u, nil
}

func (s *Store) nextSortOrder(ctx context.Context, tx pgx.Tx, userID int64, at time.Time) (int, error) {
	const q = `
SELECT COALESCE(MAX(sort_order), -1) + 1
FROM meals
WHERE user_id = $1
  AND (created_at AT TIME ZONE 'Europe/Kyiv')::date = ($2 AT TIME ZONE 'Europe/Kyiv')::date`
	var order int
	if err := tx.QueryRow(ctx, q, userID, at.UTC()).Scan(&order); err != nil {
		return 0, fmt.Errorf("next sort order: %w", err)
	}
	return order, nil
}

func (s *Store) CreateMeal(ctx context.Context, userID int64, rawText string, parsed domain.ParsedMeal, createdAt time.Time) (domain.Meal, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Meal{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	} else {
		createdAt = createdAt.UTC()
	}

	sortOrder, err := s.nextSortOrder(ctx, tx, userID, createdAt)
	if err != nil {
		return domain.Meal{}, err
	}

	const insertMeal = `
INSERT INTO meals (user_id, total_calories, total_protein, total_fat, total_carbs, raw_text, sort_order, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, user_id, total_calories, total_protein, total_fat, total_carbs, raw_text, sort_order, created_at`

	var meal domain.Meal
	err = tx.QueryRow(
		ctx,
		insertMeal,
		userID,
		parsed.TotalCalories,
		parsed.TotalProtein,
		parsed.TotalFat,
		parsed.TotalCarbs,
		rawText,
		sortOrder,
		createdAt,
	).Scan(
		&meal.ID,
		&meal.UserID,
		&meal.TotalCalories,
		&meal.TotalProtein,
		&meal.TotalFat,
		&meal.TotalCarbs,
		&meal.RawText,
		&meal.SortOrder,
		&meal.CreatedAt,
	)
	if err != nil {
		return domain.Meal{}, fmt.Errorf("insert meal: %w", err)
	}

	items, err := s.insertMealItems(ctx, tx, meal.ID, parsed.Items)
	if err != nil {
		return domain.Meal{}, err
	}
	meal.Items = items

	if err := tx.Commit(ctx); err != nil {
		return domain.Meal{}, fmt.Errorf("commit meal: %w", err)
	}

	return meal, nil
}

func (s *Store) insertMealItems(ctx context.Context, tx pgx.Tx, mealID int64, items []domain.ParsedItem) ([]domain.MealItem, error) {
	const insertItem = `
INSERT INTO meal_items (meal_id, name, weight_g, calories, protein, fat, carbs, category)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, meal_id, name, weight_g, calories, protein, fat, carbs, category`

	out := make([]domain.MealItem, 0, len(items))
	for _, item := range items {
		category := item.Category
		if !domain.IsValidCategory(category) {
			category = domain.CategoryHighQualityProtein
		}
		var mi domain.MealItem
		err := tx.QueryRow(
			ctx,
			insertItem,
			mealID,
			item.Name,
			item.WeightG,
			item.Calories,
			item.Protein,
			item.Fat,
			item.Carbs,
			category,
		).Scan(
			&mi.ID,
			&mi.MealID,
			&mi.Name,
			&mi.WeightG,
			&mi.Calories,
			&mi.Protein,
			&mi.Fat,
			&mi.Carbs,
			&mi.Category,
		)
		if err != nil {
			return nil, fmt.Errorf("insert meal item: %w", err)
		}
		out = append(out, mi)
	}
	return out, nil
}

func (s *Store) GetMeal(ctx context.Context, userID, mealID int64) (domain.Meal, error) {
	const q = `
SELECT id, user_id, total_calories, total_protein, total_fat, total_carbs, raw_text, sort_order, created_at
FROM meals
WHERE id = $1 AND user_id = $2`

	var meal domain.Meal
	err := s.pool.QueryRow(ctx, q, mealID, userID).Scan(
		&meal.ID,
		&meal.UserID,
		&meal.TotalCalories,
		&meal.TotalProtein,
		&meal.TotalFat,
		&meal.TotalCarbs,
		&meal.RawText,
		&meal.SortOrder,
		&meal.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.Meal{}, fmt.Errorf("meal not found")
		}
		return domain.Meal{}, fmt.Errorf("get meal: %w", err)
	}

	itemsByMeal, err := s.loadItems(ctx, []int64{meal.ID})
	if err != nil {
		return domain.Meal{}, err
	}
	meal.Items = itemsByMeal[meal.ID]
	if meal.Items == nil {
		meal.Items = []domain.MealItem{}
	}
	return meal, nil
}

func (s *Store) UpdateMeal(ctx context.Context, userID, mealID int64, rawText string, parsed domain.ParsedMeal, createdAt time.Time) (domain.Meal, error) {
	existing, err := s.GetMeal(ctx, userID, mealID)
	if err != nil {
		return domain.Meal{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Meal{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	createdAt = createdAt.UTC()
	sortOrder := existing.SortOrder
	oldDay := existing.CreatedAt.UTC().Format("2006-01-02")
	newDay := createdAt.Format("2006-01-02")
	if oldDay != newDay {
		sortOrder, err = s.nextSortOrder(ctx, tx, userID, createdAt)
		if err != nil {
			return domain.Meal{}, err
		}
	}

	const updateMeal = `
UPDATE meals
SET total_calories = $3,
    total_protein = $4,
    total_fat = $5,
    total_carbs = $6,
    raw_text = $7,
    sort_order = $8,
    created_at = $9
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, total_calories, total_protein, total_fat, total_carbs, raw_text, sort_order, created_at`

	var meal domain.Meal
	err = tx.QueryRow(
		ctx,
		updateMeal,
		mealID,
		userID,
		parsed.TotalCalories,
		parsed.TotalProtein,
		parsed.TotalFat,
		parsed.TotalCarbs,
		rawText,
		sortOrder,
		createdAt,
	).Scan(
		&meal.ID,
		&meal.UserID,
		&meal.TotalCalories,
		&meal.TotalProtein,
		&meal.TotalFat,
		&meal.TotalCarbs,
		&meal.RawText,
		&meal.SortOrder,
		&meal.CreatedAt,
	)
	if err != nil {
		return domain.Meal{}, fmt.Errorf("update meal: %w", err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM meal_items WHERE meal_id = $1`, mealID); err != nil {
		return domain.Meal{}, fmt.Errorf("delete meal items: %w", err)
	}

	items, err := s.insertMealItems(ctx, tx, meal.ID, parsed.Items)
	if err != nil {
		return domain.Meal{}, err
	}
	meal.Items = items

	if err := tx.Commit(ctx); err != nil {
		return domain.Meal{}, fmt.Errorf("commit update meal: %w", err)
	}
	return meal, nil
}

func (s *Store) DeleteMeal(ctx context.Context, userID, mealID int64) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM meals WHERE id = $1 AND user_id = $2`, mealID, userID)
	if err != nil {
		return fmt.Errorf("delete meal: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("meal not found")
	}
	return nil
}

func (s *Store) ReorderMeals(ctx context.Context, userID int64, day time.Time, mealIDs []int64) error {
	day = time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := day.Add(24 * time.Hour)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	const listQ = `
SELECT id
FROM meals
WHERE user_id = $1 AND created_at >= $2 AND created_at < $3
ORDER BY sort_order ASC, created_at ASC, id ASC`
	rows, err := tx.Query(ctx, listQ, userID, day, dayEnd)
	if err != nil {
		return fmt.Errorf("list day meals: %w", err)
	}
	existing := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return fmt.Errorf("scan day meal: %w", err)
		}
		existing = append(existing, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	if len(existing) != len(mealIDs) {
		return fmt.Errorf("meal_ids must include every meal for the day")
	}
	seen := make(map[int64]struct{}, len(mealIDs))
	for _, id := range mealIDs {
		seen[id] = struct{}{}
	}
	for _, id := range existing {
		if _, ok := seen[id]; !ok {
			return fmt.Errorf("meal_ids must include every meal for the day")
		}
	}

	for i, id := range mealIDs {
		tag, err := tx.Exec(ctx, `
UPDATE meals SET sort_order = $3
WHERE id = $1 AND user_id = $2 AND created_at >= $4 AND created_at < $5`, id, userID, i, day, dayEnd)
		if err != nil {
			return fmt.Errorf("reorder meal: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("meal not found for day")
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit reorder: %w", err)
	}
	return nil
}

func (s *Store) ListMeals(ctx context.Context, userID int64, from, to time.Time, limit int) ([]domain.Meal, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	const q = `
SELECT id, user_id, total_calories, total_protein, total_fat, total_carbs, raw_text, sort_order, created_at
FROM meals
WHERE user_id = $1 AND created_at >= $2 AND created_at < $3
ORDER BY (created_at AT TIME ZONE 'Europe/Kyiv')::date DESC, sort_order ASC, created_at ASC, id ASC
LIMIT $4`

	rows, err := s.pool.Query(ctx, q, userID, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("list meals: %w", err)
	}
	defer rows.Close()

	meals := make([]domain.Meal, 0)
	ids := make([]int64, 0)
	for rows.Next() {
		var m domain.Meal
		if err := rows.Scan(
			&m.ID,
			&m.UserID,
			&m.TotalCalories,
			&m.TotalProtein,
			&m.TotalFat,
			&m.TotalCarbs,
			&m.RawText,
			&m.SortOrder,
			&m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan meal: %w", err)
		}
		meals = append(meals, m)
		ids = append(ids, m.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list meals rows: %w", err)
	}

	if len(ids) == 0 {
		return meals, nil
	}

	itemsByMeal, err := s.loadItems(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range meals {
		meals[i].Items = itemsByMeal[meals[i].ID]
		if meals[i].Items == nil {
			meals[i].Items = []domain.MealItem{}
		}
	}

	return meals, nil
}

func (s *Store) loadItems(ctx context.Context, mealIDs []int64) (map[int64][]domain.MealItem, error) {
	const q = `
SELECT id, meal_id, name, weight_g, calories, protein, fat, carbs, category
FROM meal_items
WHERE meal_id = ANY($1)
ORDER BY id ASC`

	rows, err := s.pool.Query(ctx, q, mealIDs)
	if err != nil {
		return nil, fmt.Errorf("list meal items: %w", err)
	}
	defer rows.Close()

	result := make(map[int64][]domain.MealItem)
	for rows.Next() {
		var item domain.MealItem
		if err := rows.Scan(
			&item.ID,
			&item.MealID,
			&item.Name,
			&item.WeightG,
			&item.Calories,
			&item.Protein,
			&item.Fat,
			&item.Carbs,
			&item.Category,
		); err != nil {
			return nil, fmt.Errorf("scan meal item: %w", err)
		}
		result[item.MealID] = append(result[item.MealID], item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("meal items rows: %w", err)
	}
	return result, nil
}

func (s *Store) StatsSummary(ctx context.Context, userID int64, from, to time.Time) (domain.StatsSummary, error) {
	summary := domain.StatsSummary{
		From:  from,
		To:    to,
		Daily: []domain.DailyStat{},
	}

	const totalsQ = `
SELECT
  COALESCE(SUM(total_calories), 0),
  COALESCE(SUM(total_protein), 0),
  COALESCE(SUM(total_fat), 0),
  COALESCE(SUM(total_carbs), 0),
  COUNT(*)
FROM meals
WHERE user_id = $1 AND created_at >= $2 AND created_at < $3`

	err := s.pool.QueryRow(ctx, totalsQ, userID, from, to).Scan(
		&summary.TotalCalories,
		&summary.TotalProtein,
		&summary.TotalFat,
		&summary.TotalCarbs,
		&summary.MealCount,
	)
	if err != nil {
		return domain.StatsSummary{}, fmt.Errorf("stats totals: %w", err)
	}

	const dailyQ = `
SELECT
  (created_at AT TIME ZONE 'Europe/Kyiv')::date AS day,
  COALESCE(SUM(total_calories), 0),
  COALESCE(SUM(total_protein), 0),
  COALESCE(SUM(total_fat), 0),
  COALESCE(SUM(total_carbs), 0),
  COUNT(*)
FROM meals
WHERE user_id = $1 AND created_at >= $2 AND created_at < $3
GROUP BY day
ORDER BY day ASC`

	rows, err := s.pool.Query(ctx, dailyQ, userID, from, to)
	if err != nil {
		return domain.StatsSummary{}, fmt.Errorf("stats daily: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var d domain.DailyStat
		if err := rows.Scan(&d.Date, &d.Calories, &d.Protein, &d.Fat, &d.Carbs, &d.Meals); err != nil {
			return domain.StatsSummary{}, fmt.Errorf("scan daily stat: %w", err)
		}
		summary.Daily = append(summary.Daily, d)
	}
	if err := rows.Err(); err != nil {
		return domain.StatsSummary{}, fmt.Errorf("daily rows: %w", err)
	}

	return summary, nil
}

func (s *Store) FindUserByTelegramID(ctx context.Context, telegramID int64) (domain.User, error) {
	const q = `SELECT ` + userSelectColumns + ` FROM users WHERE telegram_id = $1`
	u, err := scanUser(s.pool.QueryRow(ctx, q, telegramID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, fmt.Errorf("user not found")
		}
		return domain.User{}, fmt.Errorf("find user: %w", err)
	}
	return u, nil
}

func (s *Store) FindUserByID(ctx context.Context, userID int64) (domain.User, error) {
	const q = `SELECT ` + userSelectColumns + ` FROM users WHERE id = $1`
	u, err := scanUser(s.pool.QueryRow(ctx, q, userID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, fmt.Errorf("user not found")
		}
		return domain.User{}, fmt.Errorf("find user by id: %w", err)
	}
	return u, nil
}

func (s *Store) FindUserByEmail(ctx context.Context, email string) (domain.User, error) {
	const q = `SELECT ` + userSelectColumns + ` FROM users WHERE lower(email) = lower($1)`
	u, err := scanUser(s.pool.QueryRow(ctx, q, email))
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, fmt.Errorf("user not found")
		}
		return domain.User{}, fmt.Errorf("find user by email: %w", err)
	}
	return u, nil
}

func (s *Store) SetUserCredentials(ctx context.Context, userID int64, email, passwordHash string) (domain.User, error) {
	const q = `
UPDATE users
SET email = $2, password_hash = $3
WHERE id = $1
RETURNING ` + userSelectColumns
	u, err := scanUser(s.pool.QueryRow(ctx, q, userID, email, passwordHash))
	if err != nil {
		return domain.User{}, fmt.Errorf("set user credentials: %w", err)
	}
	return u, nil
}

func (s *Store) UpdateUserProfile(ctx context.Context, userID int64, profile domain.UserProfile) (domain.User, error) {
	const q = `
UPDATE users
SET weight_kg = $2,
    height_cm = $3,
    age = $4,
    sex = NULLIF($5, ''),
    activity_level = NULLIF($6, '')
WHERE id = $1
RETURNING ` + userSelectColumns
	u, err := scanUser(s.pool.QueryRow(
		ctx,
		q,
		userID,
		profile.WeightKg,
		profile.HeightCm,
		profile.Age,
		profile.Sex,
		profile.ActivityLevel,
	))
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, fmt.Errorf("user not found")
		}
		return domain.User{}, fmt.Errorf("update user profile: %w", err)
	}
	return u, nil
}

func (s *Store) CreateLinkCode(ctx context.Context, userID int64, code string, expiresAt time.Time) (domain.LinkCode, error) {
	const cleanup = `DELETE FROM link_codes WHERE user_id = $1 OR expires_at < NOW()`
	if _, err := s.pool.Exec(ctx, cleanup, userID); err != nil {
		return domain.LinkCode{}, fmt.Errorf("cleanup link codes: %w", err)
	}

	const q = `
INSERT INTO link_codes (code, user_id, expires_at)
VALUES ($1, $2, $3)
RETURNING code, user_id, expires_at, created_at`
	var lc domain.LinkCode
	err := s.pool.QueryRow(ctx, q, code, userID, expiresAt).Scan(
		&lc.Code,
		&lc.UserID,
		&lc.ExpiresAt,
		&lc.CreatedAt,
	)
	if err != nil {
		return domain.LinkCode{}, fmt.Errorf("create link code: %w", err)
	}
	return lc, nil
}

func (s *Store) UpdateUserLanguage(ctx context.Context, userID int64, language string) (domain.User, error) {
	const q = `
UPDATE users
SET language = $2
WHERE id = $1
RETURNING ` + userSelectColumns
	u, err := scanUser(s.pool.QueryRow(ctx, q, userID, language))
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, fmt.Errorf("user not found")
		}
		return domain.User{}, fmt.Errorf("update user language: %w", err)
	}
	return u, nil
}

func (s *Store) FindLinkCodeUser(ctx context.Context, code string) (domain.User, error) {
	const q = `
SELECT u.id, u.telegram_id, COALESCE(u.email, ''), COALESCE(u.password_hash, ''), u.created_at,
       u.weight_kg, u.height_cm, u.age, COALESCE(u.sex, ''), COALESCE(u.activity_level, ''), COALESCE(u.language, '')
FROM link_codes lc
JOIN users u ON u.id = lc.user_id
WHERE lc.code = $1 AND lc.expires_at > NOW()`
	u, err := scanUser(s.pool.QueryRow(ctx, q, code))
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, fmt.Errorf("link code not found or expired")
		}
		return domain.User{}, fmt.Errorf("find link code user: %w", err)
	}
	return u, nil
}

func (s *Store) DeleteLinkCode(ctx context.Context, code string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM link_codes WHERE code = $1`, code)
	if err != nil {
		return fmt.Errorf("delete link code: %w", err)
	}
	return nil
}
