package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/root1/calories-tracker/internal/domain"
	"github.com/root1/calories-tracker/internal/parser"
)

var (
	ErrMealNotFound = errors.New("meal not found")
	ErrInvalidMeal  = errors.New("invalid meal")
)

type MealStore interface {
	EnsureUser(ctx context.Context, telegramID int64) (domain.User, error)
	CreateMeal(ctx context.Context, userID int64, rawText string, parsed domain.ParsedMeal, createdAt time.Time) (domain.Meal, error)
	GetMeal(ctx context.Context, userID, mealID int64) (domain.Meal, error)
	UpdateMeal(ctx context.Context, userID, mealID int64, rawText string, parsed domain.ParsedMeal, createdAt time.Time) (domain.Meal, error)
	DeleteMeal(ctx context.Context, userID, mealID int64) error
	ReorderMeals(ctx context.Context, userID int64, day time.Time, mealIDs []int64) error
	ListMeals(ctx context.Context, userID int64, from, to time.Time, limit int) ([]domain.Meal, error)
	StatsSummary(ctx context.Context, userID int64, from, to time.Time) (domain.StatsSummary, error)
	FindUserByID(ctx context.Context, userID int64) (domain.User, error)
}

type MealService struct {
	store  MealStore
	parser parser.MealParser
}

type CreateMealInput struct {
	RawText   string
	CreatedAt *time.Time
	Items     []domain.ParsedItem
}

type UpdateMealInput struct {
	RawText   *string
	CreatedAt *time.Time
	Items     []domain.ParsedItem
}

func NewMealService(store MealStore, mealParser parser.MealParser) *MealService {
	return &MealService{store: store, parser: mealParser}
}

func (s *MealService) LogMealFromText(ctx context.Context, telegramID int64, rawText string) (domain.Meal, error) {
	parsed, err := s.parser.Parse(ctx, rawText)
	if err != nil {
		return domain.Meal{}, fmt.Errorf("parse meal: %w", err)
	}

	user, err := s.store.EnsureUser(ctx, telegramID)
	if err != nil {
		return domain.Meal{}, err
	}

	createdAt, err := parser.ResolveEatenAt(parsed.EatenAt, time.Now().UTC())
	if err != nil {
		return domain.Meal{}, fmt.Errorf("parse meal: %w", err)
	}

	meal, err := s.store.CreateMeal(ctx, user.ID, rawText, parsed, createdAt)
	if err != nil {
		return domain.Meal{}, err
	}
	return meal, nil
}

func (s *MealService) GetStats(ctx context.Context, userID int64, from, to time.Time) (domain.StatsSummary, error) {
	if _, err := s.store.FindUserByID(ctx, userID); err != nil {
		return domain.StatsSummary{}, err
	}
	return s.store.StatsSummary(ctx, userID, from, to)
}

func (s *MealService) ListMeals(ctx context.Context, userID int64, from, to time.Time, limit int) ([]domain.Meal, error) {
	if _, err := s.store.FindUserByID(ctx, userID); err != nil {
		return nil, err
	}
	return s.store.ListMeals(ctx, userID, from, to, limit)
}

func (s *MealService) CreateMealForUser(ctx context.Context, userID int64, input CreateMealInput) (domain.Meal, error) {
	if _, err := s.store.FindUserByID(ctx, userID); err != nil {
		return domain.Meal{}, err
	}

	parsed, rawText, err := s.resolveParsedMeal(ctx, input.RawText, input.Items)
	if err != nil {
		return domain.Meal{}, err
	}

	createdAt, err := resolveMealCreatedAt(input.CreatedAt, parsed)
	if err != nil {
		return domain.Meal{}, err
	}

	meal, err := s.store.CreateMeal(ctx, userID, rawText, parsed, createdAt)
	if err != nil {
		return domain.Meal{}, err
	}
	return meal, nil
}

func (s *MealService) UpdateMealForUser(ctx context.Context, userID, mealID int64, input UpdateMealInput) (domain.Meal, error) {
	existing, err := s.store.GetMeal(ctx, userID, mealID)
	if err != nil {
		return domain.Meal{}, ErrMealNotFound
	}

	rawText := existing.RawText
	if input.RawText != nil {
		rawText = *input.RawText
	}

	var parsed domain.ParsedMeal
	if len(input.Items) > 0 {
		parsed, err = validateManualItems(input.Items)
		if err != nil {
			return domain.Meal{}, err
		}
	} else if input.RawText != nil && strings.TrimSpace(*input.RawText) != "" {
		parsed, err = s.parser.Parse(ctx, *input.RawText)
		if err != nil {
			return domain.Meal{}, fmt.Errorf("%w: %v", ErrInvalidMeal, err)
		}
	} else {
		parsed = mealToParsed(existing)
	}

	createdAt, err := resolveMealCreatedAt(input.CreatedAt, parsed)
	if err != nil {
		return domain.Meal{}, err
	}
	if createdAt.IsZero() {
		createdAt = existing.CreatedAt
	}

	meal, err := s.store.UpdateMeal(ctx, userID, mealID, rawText, parsed, createdAt)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return domain.Meal{}, ErrMealNotFound
		}
		return domain.Meal{}, err
	}
	return meal, nil
}

func (s *MealService) DeleteMealForUser(ctx context.Context, userID, mealID int64) error {
	if err := s.store.DeleteMeal(ctx, userID, mealID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return ErrMealNotFound
		}
		return err
	}
	return nil
}

func (s *MealService) ReorderMealsForUser(ctx context.Context, userID int64, day time.Time, mealIDs []int64) error {
	if len(mealIDs) == 0 {
		return fmt.Errorf("%w: empty meal_ids", ErrInvalidMeal)
	}
	if err := s.store.ReorderMeals(ctx, userID, day, mealIDs); err != nil {
		if strings.Contains(err.Error(), "must include") || strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("%w: %v", ErrInvalidMeal, err)
		}
		return err
	}
	return nil
}

func (s *MealService) resolveParsedMeal(ctx context.Context, rawText string, items []domain.ParsedItem) (domain.ParsedMeal, string, error) {
	rawText = strings.TrimSpace(rawText)
	if len(items) > 0 {
		parsed, err := validateManualItems(items)
		if err != nil {
			return domain.ParsedMeal{}, "", err
		}
		return parsed, rawText, nil
	}
	if rawText == "" {
		return domain.ParsedMeal{}, "", fmt.Errorf("%w: raw_text or items required", ErrInvalidMeal)
	}
	parsed, err := s.parser.Parse(ctx, rawText)
	if err != nil {
		return domain.ParsedMeal{}, "", fmt.Errorf("%w: %v", ErrInvalidMeal, err)
	}
	return parsed, rawText, nil
}

func resolveMealCreatedAt(explicit *time.Time, parsed domain.ParsedMeal) (time.Time, error) {
	if explicit != nil {
		return explicit.UTC(), nil
	}
	if strings.TrimSpace(parsed.EatenAt) != "" {
		return parser.ResolveEatenAt(parsed.EatenAt, time.Now().UTC())
	}
	return time.Time{}, nil
}

func validateManualItems(items []domain.ParsedItem) (domain.ParsedMeal, error) {
	parsed := domain.ParsedMeal{Items: items}
	return parser.ValidateParsedMeal(parsed.RecalcTotals())
}

func mealToParsed(meal domain.Meal) domain.ParsedMeal {
	items := make([]domain.ParsedItem, 0, len(meal.Items))
	for _, item := range meal.Items {
		items = append(items, domain.ParsedItem{
			Name:     item.Name,
			WeightG:  item.WeightG,
			Calories: item.Calories,
			Protein:  item.Protein,
			Fat:      item.Fat,
			Carbs:    item.Carbs,
			Category: item.Category,
		})
	}
	return domain.ParsedMeal{
		TotalCalories: meal.TotalCalories,
		TotalProtein:  meal.TotalProtein,
		TotalFat:      meal.TotalFat,
		TotalCarbs:    meal.TotalCarbs,
		EatenAt:       meal.CreatedAt.In(parser.MealLocation()).Format("2006-01-02"),
		Items:         items,
	}
}

func FormatMealReply(meal domain.Meal, lang string) string {
	lang = domain.NormalizeLanguage(lang)
	dateLabel := meal.CreatedAt.In(parser.MealLocation()).Format("02.01.2006")
	if lang == domain.LangEN {
		msg := fmt.Sprintf(
			"Saved ✅ (%s)\nCalories: %.0f\nProtein: %.1f g\nFat: %.1f g\nCarbs: %.1f g\n\nItems:\n",
			dateLabel,
			meal.TotalCalories,
			meal.TotalProtein,
			meal.TotalFat,
			meal.TotalCarbs,
		)
		for _, item := range meal.Items {
			msg += fmt.Sprintf("• %s (%.0f g) — %.0f kcal [%s]\n", item.Name, item.WeightG, item.Calories, categoryLabel(item.Category, lang))
		}
		return msg
	}

	msg := fmt.Sprintf(
		"Записано ✅ (%s)\nКалорії: %.0f\nБілки: %.1f г\nЖири: %.1f г\nВуглеводи: %.1f г\n\nСтрави:\n",
		dateLabel,
		meal.TotalCalories,
		meal.TotalProtein,
		meal.TotalFat,
		meal.TotalCarbs,
	)
	for _, item := range meal.Items {
		msg += fmt.Sprintf("• %s (%.0f г) — %.0f ккал [%s]\n", item.Name, item.WeightG, item.Calories, categoryLabel(item.Category, lang))
	}
	return msg
}

func categoryLabel(category, lang string) string {
	if lang == domain.LangEN {
		switch category {
		case domain.CategoryHighQualityProtein:
			return "protein"
		case domain.CategoryLongActingCarbs:
			return "slow carbs"
		case domain.CategoryLipids:
			return "fats"
		case domain.CategoryFastActingCarbs:
			return "fast carbs"
		default:
			return category
		}
	}
	switch category {
	case domain.CategoryHighQualityProtein:
		return "білок"
	case domain.CategoryLongActingCarbs:
		return "повільні вугл."
	case domain.CategoryLipids:
		return "жири"
	case domain.CategoryFastActingCarbs:
		return "швидкі вугл."
	default:
		return category
	}
}
