package service

import (
	"strings"
	"testing"
	"time"

	"github.com/root1/calories-tracker/internal/domain"
)

func TestFormatMealReply(t *testing.T) {
	meal := domain.Meal{
		TotalCalories: 500,
		TotalProtein:  30,
		TotalFat:      20,
		TotalCarbs:    40,
		CreatedAt:     time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC),
		Items: []domain.MealItem{
			{Name: "Вівсянка", WeightG: 300, Calories: 300},
			{Name: "Яйце", WeightG: 50, Calories: 200},
		},
	}

	msg := FormatMealReply(meal, domain.LangUK)
	if !strings.Contains(msg, "Записано") {
		t.Fatalf("expected success header, got %q", msg)
	}
	if !strings.Contains(msg, "05.08.2026") {
		t.Fatalf("expected date in reply, got %q", msg)
	}
	if !strings.Contains(msg, "Вівсянка") || !strings.Contains(msg, "Яйце") {
		t.Fatalf("expected item names in reply, got %q", msg)
	}

	en := FormatMealReply(meal, domain.LangEN)
	if !strings.Contains(en, "Saved") {
		t.Fatalf("expected english header, got %q", en)
	}
}
