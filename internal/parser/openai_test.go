package parser

import (
	"testing"
)

func TestDecodeAndValidate_ValidUkrainianMeal(t *testing.T) {
	content := `{
  "total_calories": 650,
  "total_protein": 45,
  "total_fat": 22,
  "total_carbs": 55,
  "eaten_at": "2026-08-05",
  "items": [
    {"name": "Вівсянка з молоком", "weight_g": 350, "calories": 325, "protein": 12, "fat": 8, "carbs": 50, "category": "long_acting_carbs"},
    {"name": "Яєчня з 2 яєць", "weight_g": 120, "calories": 180, "protein": 14, "fat": 13, "carbs": 1, "category": "high_quality_protein"},
    {"name": "Банан", "weight_g": 120, "calories": 145, "protein": 1.5, "fat": 0.5, "carbs": 4, "category": "fast_acting_carbs"}
  ]
}`

	meal, err := DecodeAndValidate(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(meal.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(meal.Items))
	}
	if meal.Items[0].Name != "Вівсянка з молоком" {
		t.Fatalf("unexpected first item name: %s", meal.Items[0].Name)
	}
	if meal.Items[0].Category != "long_acting_carbs" {
		t.Fatalf("unexpected category: %s", meal.Items[0].Category)
	}
	if meal.TotalCalories != 650 {
		t.Fatalf("unexpected total calories: %v", meal.TotalCalories)
	}
	if meal.EatenAt != "2026-08-05" {
		t.Fatalf("unexpected eaten_at: %s", meal.EatenAt)
	}
}

func TestDecodeAndValidate_StripsMarkdownFence(t *testing.T) {
	content := "```json\n{\"total_calories\":100,\"total_protein\":10,\"total_fat\":5,\"total_carbs\":8,\"items\":[{\"name\":\"Хліб\",\"weight_g\":40,\"calories\":100,\"protein\":10,\"fat\":5,\"carbs\":8,\"category\":\"fast_acting_carbs\"}]}\n```"

	meal, err := DecodeAndValidate(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meal.Items[0].Name != "Хліб" {
		t.Fatalf("unexpected name: %s", meal.Items[0].Name)
	}
}

func TestDecodeAndValidate_EmptyItems(t *testing.T) {
	content := `{"total_calories":0,"total_protein":0,"total_fat":0,"total_carbs":0,"items":[]}`
	_, err := DecodeAndValidate(content)
	if err == nil {
		t.Fatal("expected error for empty items")
	}
}

func TestDecodeAndValidate_NegativeValues(t *testing.T) {
	content := `{"total_calories":-1,"total_protein":0,"total_fat":0,"total_carbs":0,"items":[{"name":"X","weight_g":10,"calories":10,"protein":1,"fat":1,"carbs":1,"category":"lipids"}]}`
	_, err := DecodeAndValidate(content)
	if err == nil {
		t.Fatal("expected error for negative totals")
	}
}

func TestDecodeAndValidate_EmptyName(t *testing.T) {
	content := `{"total_calories":10,"total_protein":1,"total_fat":1,"total_carbs":1,"items":[{"name":"","weight_g":10,"calories":10,"protein":1,"fat":1,"carbs":1,"category":"lipids"}]}`
	_, err := DecodeAndValidate(content)
	if err == nil {
		t.Fatal("expected error for empty item name")
	}
}

func TestDecodeAndValidate_InvalidCategory(t *testing.T) {
	content := `{"total_calories":10,"total_protein":1,"total_fat":1,"total_carbs":1,"items":[{"name":"X","weight_g":10,"calories":10,"protein":1,"fat":1,"carbs":1,"category":"junk"}]}`
	_, err := DecodeAndValidate(content)
	if err == nil {
		t.Fatal("expected error for invalid category")
	}
}
