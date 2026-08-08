package nutrition

import (
	"testing"

	"github.com/root1/calories-tracker/internal/domain"
)

func TestCalculateDailyTargets(t *testing.T) {
	weight := 80.0
	height := 180.0
	age := 30

	targets, err := CalculateDailyTargets(domain.UserProfile{
		WeightKg:      &weight,
		HeightCm:      &height,
		Age:           &age,
		Sex:           domain.SexMale,
		ActivityLevel: domain.ActivityModerate,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if targets.BMR < 1700 || targets.BMR > 1850 {
		t.Fatalf("unexpected bmr: %v", targets.BMR)
	}
	if targets.Calories < targets.BMR {
		t.Fatalf("tdee should exceed bmr: calories=%v bmr=%v", targets.Calories, targets.BMR)
	}
	if targets.Protein < 140 || targets.Protein > 150 {
		t.Fatalf("unexpected protein: %v", targets.Protein)
	}
	if targets.Fat <= 0 || targets.Carbs <= 0 {
		t.Fatalf("fat and carbs should be positive: fat=%v carbs=%v", targets.Fat, targets.Carbs)
	}
}

func TestCalculateDailyTargetsRequiresFields(t *testing.T) {
	_, err := CalculateDailyTargets(domain.UserProfile{})
	if err == nil {
		t.Fatal("expected error for empty profile")
	}
}
