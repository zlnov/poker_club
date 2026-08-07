package telegram

import (
	"context"
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"poker-club/backend/internal/config"
	"poker-club/backend/internal/service"
)

// Bot wraps the Telegram Bot API client and provides update processing.
type Bot struct {
	api *tgbotapi.BotAPI
	svc *service.Service
	log *slog.Logger
}

// New creates a new Telegram Bot instance.
func New(cfg *config.Config, svc *service.Service, log *slog.Logger) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}
	api.Debug = false

	log.Info("telegram bot authorized", "username", api.Self.UserName)

	return &Bot{
		api: api,
		svc: svc,
		log: log,
	}, nil
}

// API returns the underlying BotAPI instance.
func (b *Bot) API() *tgbotapi.BotAPI {
	return b.api
}

// ProcessUpdate handles a single Telegram update.
func (b *Bot) ProcessUpdate(update tgbotapi.Update) {
	b.log.Info("received update", "update_id", update.UpdateID)
}

// SetupWebhook registers the webhook URL with Telegram.
func (b *Bot) SetupWebhook(ctx context.Context, webhookURL string) error {
	wh, err := tgbotapi.NewWebhook(webhookURL)
	if err != nil {
		return fmt.Errorf("failed to create webhook config: %w", err)
	}
	_, err = b.api.Request(wh)
	if err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}
	b.log.Info("webhook set", "url", webhookURL)
	return nil
}

// StartLongPolling starts receiving updates via long polling.
// This mode is intended for local development only.
func (b *Bot) StartLongPolling(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)
	b.log.Info("long polling started")

	for {
		select {
		case <-ctx.Done():
			b.log.Info("long polling stopped")
			return nil
		case update := <-updates:
			b.ProcessUpdate(update)
		}
	}
}
