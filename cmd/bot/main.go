package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/root1/calories-tracker/internal/config"
	"github.com/root1/calories-tracker/internal/httpapi"
	"github.com/root1/calories-tracker/internal/parser"
	"github.com/root1/calories-tracker/internal/service"
	"github.com/root1/calories-tracker/internal/storage/postgres"
	"github.com/root1/calories-tracker/internal/storage/postgres/migrate"
	"github.com/root1/calories-tracker/internal/telegram"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = filepath.Join(".", "migrations")
	}
	if err := migrate.Up(cfg.DatabaseURL, migrationsPath); err != nil {
		logger.Error("migrate", "err", err)
		os.Exit(1)
	}

	store, err := postgres.NewStore(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("postgres", "err", err)
		os.Exit(1)
	}
	defer store.Close()

	mealParser := parser.NewOpenAIClient(cfg.OpenAIAPIKey, cfg.OpenAIModel, cfg.RequestTimeout)
	meals := service.NewMealService(store, mealParser)
	auth := service.NewAuthService(store, cfg.JWTSecret, 30*24*time.Hour)
	profile := service.NewProfileService(store)
	bot := telegram.NewBotClient(cfg.TelegramBotToken, cfg.RequestTimeout)
	tgHandler := telegram.NewHandler(meals, auth, bot, cfg.WebAppURL, logger)
	api := httpapi.NewAPI(meals, auth, profile, store, cfg.CORSAllowedOrigins, logger)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           api.Router(cfg.TelegramWebhookPath, tgHandler),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("server listening", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown", "err", err)
	}
}
