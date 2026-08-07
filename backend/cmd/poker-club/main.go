package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"poker-club/backend/internal/config"
	"poker-club/backend/internal/repository/postgres"
	"poker-club/backend/internal/service"
	transporthttp "poker-club/backend/internal/transport/http"
	"poker-club/backend/internal/transport/telegram"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "application error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Load configuration
	cfg := config.Load()

	// Set up logging
	log := setupLogger(cfg.LogLevel)
	log.Info("starting poker-club backend")

	// Connect to database and run migrations
	db, err := postgres.New(ctx, cfg)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	log.Info("database connected and migrated")

	// Create repositories
	repos := postgres.NewRepositories(db)

	// Create service
	svc := service.New(repos)

	// Create HTTP server
	httpServer := transporthttp.NewServer(cfg, svc)

	// Create Telegram bot (optional — app can run without a valid token)
	var bot *telegram.Bot
	if cfg.BotToken == "" || cfg.BotToken == "test" {
		log.Warn("bot token not configured, skipping telegram bot")
	} else {
		bot, err = telegram.New(cfg, svc, log)
		if err != nil {
			log.Error("failed to create telegram bot", "error", err)
			return fmt.Errorf("failed to create telegram bot: %w", err)
		}
	}

	// Register webhook handler on HTTP server (if bot is available)
	if bot != nil {
		httpServer.RegisterWebhookHandler(bot.WebhookHandler())
	}

	// Start HTTP server (serves healthcheck and webhook)
	go func() {
		log.Info("http server starting", "addr", cfg.WebhookAddr())
		if err := httpServer.Start(); err != nil {
			log.Error("http server error", "error", err)
		}
	}()

	// Set up Telegram updates: webhook or long polling
	if bot != nil {
		if cfg.LongPolling {
			log.Info("using long polling mode (development)")
			go func() {
				if err := bot.StartLongPolling(ctx); err != nil {
					log.Error("long polling error", "error", err)
				}
			}()
		} else {
			log.Info("using webhook mode")
			if err := bot.SetupWebhook(ctx, cfg.WebhookURL); err != nil {
				log.Error("failed to set webhook", "error", err)
				return fmt.Errorf("failed to set webhook: %w", err)
			}
		}
	}

	log.Info("application started")

	// Wait for interrupt signal
	<-ctx.Done()
	log.Info("shutdown signal received")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Stop(shutdownCtx); err != nil {
		log.Error("http server shutdown error", "error", err)
	}

	log.Info("application stopped")
	return nil
}

func setupLogger(level string) *slog.Logger {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl}
	handler := slog.NewTextHandler(os.Stdout, opts)
	return slog.New(handler)
}
