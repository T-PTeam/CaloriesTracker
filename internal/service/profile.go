package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/root1/calories-tracker/internal/domain"
	"github.com/root1/calories-tracker/internal/nutrition"
)

var (
	ErrInvalidProfile = errors.New("invalid profile")
)

type ProfileStore interface {
	FindUserByID(ctx context.Context, userID int64) (domain.User, error)
	UpdateUserProfile(ctx context.Context, userID int64, profile domain.UserProfile) (domain.User, error)
}

type ProfileService struct {
	store ProfileStore
}

func NewProfileService(store ProfileStore) *ProfileService {
	return &ProfileService{store: store}
}

type ProfileResponse struct {
	WeightKg      *float64             `json:"weight_kg"`
	HeightCm      *float64             `json:"height_cm"`
	Age           *int                 `json:"age"`
	Sex           string               `json:"sex"`
	ActivityLevel string               `json:"activity_level"`
	DailyTargets  *domain.DailyTargets `json:"daily_targets"`
}

func (s *ProfileService) Get(ctx context.Context, userID int64) (ProfileResponse, error) {
	user, err := s.store.FindUserByID(ctx, userID)
	if err != nil {
		return ProfileResponse{}, err
	}
	return toProfileResponse(user), nil
}

func (s *ProfileService) Update(ctx context.Context, userID int64, profile domain.UserProfile) (ProfileResponse, error) {
	profile, err := normalizeProfile(profile)
	if err != nil {
		return ProfileResponse{}, err
	}
	user, err := s.store.UpdateUserProfile(ctx, userID, profile)
	if err != nil {
		return ProfileResponse{}, err
	}
	return toProfileResponse(user), nil
}

func toProfileResponse(user domain.User) ProfileResponse {
	resp := ProfileResponse{
		WeightKg:      user.WeightKg,
		HeightCm:      user.HeightCm,
		Age:           user.Age,
		Sex:           user.Sex,
		ActivityLevel: user.ActivityLevel,
	}
	if user.Profile().IsComplete() {
		targets, err := nutrition.CalculateDailyTargets(user.Profile())
		if err == nil {
			resp.DailyTargets = &targets
		}
	}
	return resp
}

func normalizeProfile(profile domain.UserProfile) (domain.UserProfile, error) {
	profile.Sex = strings.TrimSpace(strings.ToLower(profile.Sex))
	profile.ActivityLevel = strings.TrimSpace(strings.ToLower(profile.ActivityLevel))

	if profile.WeightKg == nil || profile.HeightCm == nil || profile.Age == nil {
		return domain.UserProfile{}, fmt.Errorf("%w: weight, height and age are required", ErrInvalidProfile)
	}
	if *profile.WeightKg <= 0 || *profile.WeightKg >= 500 {
		return domain.UserProfile{}, fmt.Errorf("%w: weight must be between 0 and 500 kg", ErrInvalidProfile)
	}
	if *profile.HeightCm <= 0 || *profile.HeightCm >= 300 {
		return domain.UserProfile{}, fmt.Errorf("%w: height must be between 0 and 300 cm", ErrInvalidProfile)
	}
	if *profile.Age < 10 || *profile.Age > 120 {
		return domain.UserProfile{}, fmt.Errorf("%w: age must be between 10 and 120", ErrInvalidProfile)
	}
	if !domain.IsValidSex(profile.Sex) {
		return domain.UserProfile{}, fmt.Errorf("%w: sex must be male or female", ErrInvalidProfile)
	}
	if !domain.IsValidActivityLevel(profile.ActivityLevel) {
		return domain.UserProfile{}, fmt.Errorf("%w: invalid activity level", ErrInvalidProfile)
	}
	return profile, nil
}
