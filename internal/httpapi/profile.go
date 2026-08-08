package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/root1/calories-tracker/internal/domain"
	"github.com/root1/calories-tracker/internal/service"
)

func (a *API) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	profile, err := a.profile.Get(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load profile")
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (a *API) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body struct {
		WeightKg      *float64 `json:"weight_kg"`
		HeightCm      *float64 `json:"height_cm"`
		Age           *int     `json:"age"`
		Sex           string   `json:"sex"`
		ActivityLevel string   `json:"activity_level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	profile, err := a.profile.Update(r.Context(), userID, domain.UserProfile{
		WeightKg:      body.WeightKg,
		HeightCm:      body.HeightCm,
		Age:           body.Age,
		Sex:           body.Sex,
		ActivityLevel: body.ActivityLevel,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidProfile) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update profile")
		return
	}
	writeJSON(w, http.StatusOK, profile)
}
