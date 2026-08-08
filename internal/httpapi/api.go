package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/root1/calories-tracker/internal/service"
)

type contextKey string

const userIDKey contextKey = "userID"

type HealthChecker interface {
	Ping(ctx context.Context) error
}

type API struct {
	meals              *service.MealService
	auth               *service.AuthService
	profile            *service.ProfileService
	health             HealthChecker
	corsAllowedOrigins []string
	logger             *slog.Logger
}

func NewAPI(
	meals *service.MealService,
	auth *service.AuthService,
	profile *service.ProfileService,
	health HealthChecker,
	corsAllowedOrigins []string,
	logger *slog.Logger,
) *API {
	if len(corsAllowedOrigins) == 0 {
		corsAllowedOrigins = []string{"http://localhost:5173", "http://127.0.0.1:5173"}
	}
	return &API{
		meals:              meals,
		auth:               auth,
		profile:            profile,
		health:             health,
		corsAllowedOrigins: corsAllowedOrigins,
		logger:             logger,
	}
}

func (a *API) Router(telegramWebhookPath string, telegramHandler http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(a.logRequests)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   a.corsAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/api/health", a.handleHealth)
	r.Post(telegramWebhookPath, telegramHandler.ServeHTTP)

	r.Post("/api/auth/register", a.handleRegister)
	r.Post("/api/auth/login", a.handleLogin)

	r.Group(func(pr chi.Router) {
		pr.Use(a.requireAuth)
		pr.Get("/api/auth/me", a.handleMe)
		pr.Patch("/api/auth/language", a.handleSetLanguage)
		pr.Get("/api/profile", a.handleGetProfile)
		pr.Patch("/api/profile", a.handleUpdateProfile)
		pr.Get("/api/stats/summary", a.handleStatsSummary)
		pr.Get("/api/meals", a.handleListMeals)
		pr.Post("/api/meals", a.handleCreateMeal)
		pr.Patch("/api/meals/reorder", a.handleReorderMeals)
		pr.Patch("/api/meals/{id}", a.handleUpdateMeal)
		pr.Delete("/api/meals/{id}", a.handleDeleteMeal)
	})

	return r
}

func (a *API) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		a.logger.Info("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", middleware.GetReqID(r.Context()),
		)
	})
}

func (a *API) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		claims, err := a.auth.ParseToken(strings.TrimSpace(parts[1]))
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func userIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(userIDKey).(int64)
	return id, ok
}

func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := a.health.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "degraded",
			"error":  "database unavailable",
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	LinkCode string `json:"link_code"`
}

func (a *API) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	result, err := a.auth.Register(r.Context(), req.Email, req.Password, req.LinkCode)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (a *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	result, err := a.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *API) handleMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	user, err := a.auth.GetPublicUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (a *API) handleSetLanguage(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body struct {
		Language string `json:"language"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	user, err := a.auth.SetLanguageByUserID(r.Context(), userID, body.Language)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid language")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (a *API) handleStatsSummary(w http.ResponseWriter, r *http.Request) {
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

	summary, err := a.meals.GetStats(r.Context(), userID, from, to)
	if err != nil {
		a.logger.Error("stats summary", "err", err)
		writeError(w, http.StatusNotFound, "user or stats not found")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func parseDateRange(r *http.Request) (from, to time.Time, err error) {
	now := time.Now().UTC()
	to = now
	from = now.AddDate(0, 0, -7)

	if raw := r.URL.Query().Get("from"); raw != "" {
		from, err = time.Parse(time.RFC3339, raw)
		if err != nil {
			from, err = time.Parse("2006-01-02", raw)
			if err != nil {
				return time.Time{}, time.Time{}, errBadRequest("invalid from")
			}
		}
	}
	if raw := r.URL.Query().Get("to"); raw != "" {
		to, err = time.Parse(time.RFC3339, raw)
		if err != nil {
			to, err = time.Parse("2006-01-02", raw)
			if err != nil {
				return time.Time{}, time.Time{}, errBadRequest("invalid to")
			}
			to = to.Add(24 * time.Hour)
		}
	}
	if !from.Before(to) {
		return time.Time{}, time.Time{}, errBadRequest("from must be before to")
	}

	return from.UTC(), to.UTC(), nil
}

func writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "invalid email or password")
	case errors.Is(err, service.ErrAlreadyRegistered):
		writeError(w, http.StatusConflict, "account already registered")
	case errors.Is(err, service.ErrInvalidLinkCode):
		writeError(w, http.StatusBadRequest, "invalid or expired link code")
	case errors.Is(err, service.ErrWeakPassword):
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
	case errors.Is(err, service.ErrInvalidEmail):
		writeError(w, http.StatusBadRequest, "invalid email")
	default:
		writeError(w, http.StatusInternalServerError, "auth failed")
	}
}

type badRequestError string

func errBadRequest(msg string) error { return badRequestError(msg) }

func (e badRequestError) Error() string { return string(e) }

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
