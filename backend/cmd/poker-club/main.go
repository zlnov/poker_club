package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"poker-club/backend/internal/auth"
	"poker-club/backend/internal/config"
	"poker-club/backend/internal/demo"
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

	cfg := config.Load()
	log := setupLogger(cfg.LogLevel)
	log.Info("starting poker-club backend")

	db, err := postgres.New(ctx, cfg)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	log.Info("database connected and migrated")

	repos := postgres.NewRepositories(db)
	svc := service.New(repos, log)

	jwtManager := auth.NewJWTManager(
		cfg.JWTSecret,
		cfg.JWTIssuer,
		cfg.JWTAudience,
		70*time.Hour,
		15*24*time.Hour,
	)

	var oidcValidator *auth.TelegramOIDCValidator
	var challengeMgr *auth.ChallengeManager
	if cfg.TelegramLoginClientID != "" {
		oidcValidator = auth.NewTelegramOIDCValidator(
			cfg.TelegramLoginClientID,
			cfg.TelegramLoginIssuer,
			cfg.TelegramLoginJWKSURL,
		)
		challengeMgr = auth.NewChallengeManager(
			cfg.JWTSecret,
			cfg.TelegramLoginNonceTTL,
			cfg.RegistrationTokenTTL,
		)
		log.Info("telegram web login (OIDC) configured")
	} else {
		log.Warn("TELEGRAM_LOGIN_CLIENT_ID not set; web telegram login disabled")
	}

	authUC := auth.NewAuthUseCase(repos, jwtManager, cfg.BotToken, log, oidcValidator, challengeMgr)

	// Optional demo database — failure must not stop production.
	var demoSvc *demo.Service
	var demoAppSvc *service.Service
	var demoDB *postgres.DB
	if cfg.DemoEnabled {
		demoDB, err = postgres.NewDemo(ctx, cfg)
		if err != nil {
			log.Warn("demo database unavailable; continuing without demo mode",
				"error", err,
				"demo_db_name", cfg.DemoDBName,
			)
		} else {
			demoRepos := postgres.NewRepositories(demoDB)
			demoSvc = demo.NewService(demoDB, demoRepos, log)
			demoAppSvc = service.New(demoRepos, log)
			// Intentionally do NOT call demoAppSvc.SetNotifier — no Telegram in demo.
			log.Info("demo database connected and migrated", "demo_db_name", cfg.DemoDBName)
			if err := demoSvc.EnsureTemplate(ctx); err != nil {
				log.Warn("demo template unavailable; demo sessions will fail until seed succeeds",
					"error", err,
				)
			}
			go demoSvc.StartCleanupLoop(ctx)
		}
	} else {
		log.Info("demo mode disabled (DEMO_ENABLED=false)")
	}
	if demoDB != nil {
		defer demoDB.Close()
	}

	httpServer := transporthttp.NewServer(cfg, svc, jwtManager, authUC, demoSvc, demoAppSvc)

	var bot *telegram.Bot
	if cfg.BotToken == "" || cfg.BotToken == "test" {
		log.Warn("bot token not configured, skipping telegram bot")
	} else {
		bot, err = telegram.New(cfg, svc, log)
		if err != nil {
			log.Error("failed to create telegram bot", "error", err)
			return fmt.Errorf("failed to create telegram bot: %w", err)
		}
		svc.SetNotifier(bot)
	}

	if bot != nil {
		httpServer.RegisterWebhookHandler(bot.WebhookHandler())
	}

	go func() {
		log.Info("http server starting", "addr", cfg.WebhookAddr())
		if err := httpServer.Start(); err != nil {
			log.Error("http server error", "error", err)
		}
	}()

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
			go bot.StartTimerProcess(ctx)
		}
	}

	log.Info("application started")

	<-ctx.Done()
	log.Info("shutdown signal received")

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
