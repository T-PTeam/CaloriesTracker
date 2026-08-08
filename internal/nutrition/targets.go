package nutrition

import (
	"fmt"
	"math"

	"github.com/root1/calories-tracker/internal/domain"
)

var activityFactors = map[string]float64{
	domain.ActivitySedentary:  1.2,
	domain.ActivityLight:      1.375,
	domain.ActivityModerate:   1.55,
	domain.ActivityActive:     1.725,
	domain.ActivityVeryActive: 1.9,
}

var proteinPerKg = map[string]float64{
	domain.ActivitySedentary:  1.4,
	domain.ActivityLight:      1.6,
	domain.ActivityModerate:   1.8,
	domain.ActivityActive:     2.0,
	domain.ActivityVeryActive: 2.2,
}

func CalculateDailyTargets(profile domain.UserProfile) (domain.DailyTargets, error) {
	if profile.WeightKg == nil || profile.HeightCm == nil || profile.Age == nil {
		return domain.DailyTargets{}, fmt.Errorf("weight, height and age are required")
	}
	if !domain.IsValidSex(profile.Sex) {
		return domain.DailyTargets{}, fmt.Errorf("sex must be male or female")
	}
	if !domain.IsValidActivityLevel(profile.ActivityLevel) {
		return domain.DailyTargets{}, fmt.Errorf("invalid activity level")
	}

	weight := *profile.WeightKg
	height := *profile.HeightCm
	age := float64(*profile.Age)

	bmr := 10*weight + 6.25*height - 5*age
	if profile.Sex == domain.SexMale {
		bmr += 5
	} else {
		bmr -= 161
	}

	tdee := bmr * activityFactors[profile.ActivityLevel]
	protein := weight * proteinPerKg[profile.ActivityLevel]
	fat := tdee * 0.30 / 9
	carbs := (tdee - protein*4 - fat*9) / 4
	if carbs < 0 {
		carbs = 0
	}

	return domain.DailyTargets{
		Calories: round1(tdee),
		Protein:  round1(protein),
		Fat:      round1(fat),
		Carbs:    round1(carbs),
		BMR:      round1(bmr),
		TDEE:     round1(tdee),
	}, nil
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}
