package main

import (
	"example.com/internal/config"
	"example.com/internal/storage/postgres"
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "production"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("server running", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	// TODO: init router: chi, chi-render

	storage, err := postgres.New(cfg.StoragePath)
	if err != nil {
		log.Error("error creating storage", err)
		os.Exit(1)
	}

	_ = storage

	// TODO: run server
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout,
				&slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout,
				&slog.HandlerOptions{Level: slog.LevelInfo}))

	}
	return log
}
