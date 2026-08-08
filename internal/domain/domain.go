package domain

import (
	"strings"
	"time"
)

const (
	CategoryHighQualityProtein = "high_quality_protein"
	CategoryLongActingCarbs    = "long_acting_carbs"
	CategoryLipids             = "lipids"
	CategoryFastActingCarbs    = "fast_acting_carbs"
)

var ValidCategories = map[string]struct{}{
	CategoryHighQualityProtein: {},
	CategoryLongActingCarbs:    {},
	CategoryLipids:             {},
	CategoryFastActingCarbs:    {},
}

func IsValidCategory(category string) bool {
	_, ok := ValidCategories[category]
	return ok
}

const (
	SexMale   = "male"
	SexFemale = "female"

	ActivitySedentary  = "sedentary"
	ActivityLight      = "light"
	ActivityModerate   = "moderate"
	ActivityActive     = "active"
	ActivityVeryActive = "very_active"
)

var ValidSexes = map[string]struct{}{
	SexMale:   {},
	SexFemale: {},
}

var ValidActivityLevels = map[string]struct{}{
	ActivitySedentary:  {},
	ActivityLight:      {},
	ActivityModerate:   {},
	ActivityActive:     {},
	ActivityVeryActive: {},
}

func IsValidSex(sex string) bool {
	_, ok := ValidSexes[sex]
	return ok
}

func IsValidActivityLevel(level string) bool {
	_, ok := ValidActivityLevels[level]
	return ok
}

const (
	LangUK = "uk"
	LangEN = "en"
)

func IsValidLanguage(lang string) bool {
	return lang == LangUK || lang == LangEN
}

func NormalizeLanguage(lang string) string {
	lang = strings.TrimSpace(strings.ToLower(lang))
	if IsValidLanguage(lang) {
		return lang
	}
	return LangUK
}

type User struct {
	ID            int64
	TelegramID    int64
	Email         string
	PasswordHash  string
	CreatedAt     time.Time
	WeightKg      *float64
	HeightCm      *float64
	Age           *int
	Sex           string
	ActivityLevel string
	Language      string
}

func (u User) HasCredentials() bool {
	return u.Email != "" && u.PasswordHash != ""
}

func (u User) Profile() UserProfile {
	return UserProfile{
		WeightKg:      u.WeightKg,
		HeightCm:      u.HeightCm,
		Age:           u.Age,
		Sex:           u.Sex,
		ActivityLevel: u.ActivityLevel,
	}
}

type UserProfile struct {
	WeightKg      *float64 `json:"weight_kg"`
	HeightCm      *float64 `json:"height_cm"`
	Age           *int     `json:"age"`
	Sex           string   `json:"sex"`
	ActivityLevel string   `json:"activity_level"`
}

func (p UserProfile) IsComplete() bool {
	return p.WeightKg != nil &&
		p.HeightCm != nil &&
		p.Age != nil &&
		IsValidSex(p.Sex) &&
		IsValidActivityLevel(p.ActivityLevel)
}

type DailyTargets struct {
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Fat      float64 `json:"fat"`
	Carbs    float64 `json:"carbs"`
	BMR      float64 `json:"bmr"`
	TDEE     float64 `json:"tdee"`
}

type LinkCode struct {
	Code      string
	UserID    int64
	ExpiresAt time.Time
	CreatedAt time.Time
}

type MealItem struct {
	ID       int64
	MealID   int64
	Name     string
	WeightG  float64
	Calories float64
	Protein  float64
	Fat      float64
	Carbs    float64
	Category string
}

type Meal struct {
	ID            int64
	UserID        int64
	TotalCalories float64
	TotalProtein  float64
	TotalFat      float64
	TotalCarbs    float64
	RawText       string
	SortOrder     int
	CreatedAt     time.Time
	Items         []MealItem
}

type ParsedMeal struct {
	TotalCalories float64      `json:"total_calories"`
	TotalProtein  float64      `json:"total_protein"`
	TotalFat      float64      `json:"total_fat"`
	TotalCarbs    float64      `json:"total_carbs"`
	EatenAt       string       `json:"eaten_at"`
	Items         []ParsedItem `json:"items"`
}

type ParsedItem struct {
	Name     string  `json:"name"`
	WeightG  float64 `json:"weight_g"`
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Fat      float64 `json:"fat"`
	Carbs    float64 `json:"carbs"`
	Category string  `json:"category"`
}

func (p ParsedMeal) RecalcTotals() ParsedMeal {
	var cal, protein, fat, carbs float64
	for _, item := range p.Items {
		cal += item.Calories
		protein += item.Protein
		fat += item.Fat
		carbs += item.Carbs
	}
	p.TotalCalories = cal
	p.TotalProtein = protein
	p.TotalFat = fat
	p.TotalCarbs = carbs
	return p
}

type DailyStat struct {
	Date     time.Time `json:"date"`
	Calories float64   `json:"calories"`
	Protein  float64   `json:"protein"`
	Fat      float64   `json:"fat"`
	Carbs    float64   `json:"carbs"`
	Meals    int       `json:"meals"`
}

type StatsSummary struct {
	From          time.Time   `json:"from"`
	To            time.Time   `json:"to"`
	TotalCalories float64     `json:"total_calories"`
	TotalProtein  float64     `json:"total_protein"`
	TotalFat      float64     `json:"total_fat"`
	TotalCarbs    float64     `json:"total_carbs"`
	MealCount     int         `json:"meal_count"`
	Daily         []DailyStat `json:"daily"`
}

type LogMealInput struct {
	TelegramID int64
	RawText    string
	Parsed     ParsedMeal
}
