package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/root1/calories-tracker/internal/domain"
	"github.com/root1/calories-tracker/internal/service"
)

type mealItemDTO struct {
	ID       int64   `json:"id,omitempty"`
	Name     string  `json:"name"`
	WeightG  float64 `json:"weight_g"`
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Fat      float64 `json:"fat"`
	Carbs    float64 `json:"carbs"`
	Category string  `json:"category"`
}

type mealDTO struct {
	ID            int64         `json:"id"`
	TotalCalories float64       `json:"total_calories"`
	TotalProtein  float64      `json:"total_protein"`
	TotalFat      float64      `json:"total_fat"`
	TotalCarbs    float64      `json:"total_carbs"`
	RawText       string        `json:"raw_text"`
	SortOrder     int           `json:"sort_order"`
	CreatedAt     time.Time     `json:"created_at"`
	Items         []mealItemDTO `json:"items"`
}

type createMealRequest struct {
	RawText   string         `json:"raw_text"`
	CreatedAt *time.Time     `json:"created_at"`
	Items     []mealItemDTO  `json:"items"`
}

type updateMealRequest struct {
	RawText   *string        `json:"raw_text"`
	CreatedAt *time.Time     `json:"created_at"`
	Items     []mealItemDTO  `json:"items"`
}

type reorderMealsRequest struct {
	Date    string  `json:"date"`
	MealIDs []int64 `json:"meal_ids"`
}

func toMealDTO(m domain.Meal) mealDTO {
	dto := mealDTO{
		ID:            m.ID,
		TotalCalories: m.TotalCalories,
		TotalProtein:  m.TotalProtein,
		TotalFat:      m.TotalFat,
		TotalCarbs:    m.TotalCarbs,
		RawText:       m.RawText,
		SortOrder:     m.SortOrder,
		CreatedAt:     m.CreatedAt,
		Items:         make([]mealItemDTO, 0, len(m.Items)),
	}
	for _, it := range m.Items {
		dto.Items = append(dto.Items, mealItemDTO{
			ID:       it.ID,
			Name:     it.Name,
			WeightG:  it.WeightG,
			Calories: it.Calories,
			Protein:  it.Protein,
			Fat:      it.Fat,
			Carbs:    it.Carbs,
			Category: it.Category,
		})
	}
	return dto
}

func itemsToParsed(items []mealItemDTO) []domain.ParsedItem {
	out := make([]domain.ParsedItem, 0, len(items))
	for _, it := range items {
		out = append(out, domain.ParsedItem{
			Name:     it.Name,
			WeightG:  it.WeightG,
			Calories: it.Calories,
			Protein:  it.Protein,
			Fat:      it.Fat,
			Carbs:    it.Carbs,
			Category: it.Category,
		})
	}
	return out
}

func (a *API) handleListMeals(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	from, to, err := parseDateRange(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			writeError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = n
	}

	meals, err := a.meals.ListMeals(r.Context(), userID, from, to, limit)
	if err != nil {
		a.logger.Error("list meals", "err", err)
		writeError(w, http.StatusNotFound, "user or meals not found")
		return
	}

	out := make([]mealDTO, 0, len(meals))
	for _, m := range meals {
		out = append(out, toMealDTO(m))
	}
	writeJSON(w, http.StatusOK, map[string]any{"meals": out})
}

func (a *API) handleCreateMeal(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createMealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	meal, err := a.meals.CreateMealForUser(r.Context(), userID, service.CreateMealInput{
		RawText:   req.RawText,
		CreatedAt: req.CreatedAt,
		Items:     itemsToParsed(req.Items),
	})
	if err != nil {
		writeMealError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toMealDTO(meal))
}

func (a *API) handleUpdateMeal(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	mealID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || mealID < 1 {
		writeError(w, http.StatusBadRequest, "invalid meal id")
		return
	}

	var req updateMealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	meal, err := a.meals.UpdateMealForUser(r.Context(), userID, mealID, service.UpdateMealInput{
		RawText:   req.RawText,
		CreatedAt: req.CreatedAt,
		Items:     itemsToParsed(req.Items),
	})
	if err != nil {
		writeMealError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toMealDTO(meal))
}

func (a *API) handleDeleteMeal(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	mealID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || mealID < 1 {
		writeError(w, http.StatusBadRequest, "invalid meal id")
		return
	}

	if err := a.meals.DeleteMealForUser(r.Context(), userID, mealID); err != nil {
		writeMealError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) handleReorderMeals(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req reorderMealsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	day, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid date")
		return
	}

	if err := a.meals.ReorderMealsForUser(r.Context(), userID, day, req.MealIDs); err != nil {
		writeMealError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeMealError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrMealNotFound):
		writeError(w, http.StatusNotFound, "meal not found")
	case errors.Is(err, service.ErrInvalidMeal):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "meal operation failed")
	}
}
