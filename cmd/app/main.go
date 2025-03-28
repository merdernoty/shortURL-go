package main

import (
	"log/slog"
	"net/http"
	"os"
	"shortURL-go/internal/config"
	save "shortURL-go/internal/http-server/handlers/url"
	mwLogger "shortURL-go/internal/http-server/middleware/logger"
	"shortURL-go/internal/storage/postgres"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	
    router.Post("/url", save.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))
	srv := &http.Server{
		Handler:      router,
		Addr:         cfg.Address,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")
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
