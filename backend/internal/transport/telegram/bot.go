package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"poker-club/backend/internal/config"
	"poker-club/backend/internal/service"
)

// userState tracks the current input state for a Telegram user.
type userState struct {
	action string // stateIdle, stateCreateClub, stateChangeName, stateCloseConfirm
	clubID int64  // relevant club ID when applicable
}

// Bot wraps the Telegram Bot API client and provides update processing.
type Bot struct {
	api    *tgbotapi.BotAPI
	svc    *service.Service
	log    *slog.Logger
	states map[int64]*userState
	mu     sync.RWMutex
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
		api:    api,
		svc:    svc,
		log:    log,
		states: make(map[int64]*userState),
	}, nil
}

// API returns the underlying BotAPI instance.
func (b *Bot) API() *tgbotapi.BotAPI {
	return b.api
}

// ProcessUpdate handles a single Telegram update.
func (b *Bot) ProcessUpdate(update tgbotapi.Update) {
	ctx := context.Background()

	// Handle callback queries (inline keyboard button presses).
	if update.CallbackQuery != nil {
		b.handleCallback(ctx, update)
		return
	}

	// Handle commands and text messages.
	if update.Message != nil {
		if update.Message.IsCommand() {
			b.handleCommand(ctx, update)
		} else {
			b.handleTextMessage(ctx, update)
		}
	}
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

// --- State management ---

// setState sets the current input state for a user.
func (b *Bot) setState(tgUserID int64, action string, clubID int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.states[tgUserID] = &userState{action: action, clubID: clubID}
}

// getState returns the current input state for a user.
func (b *Bot) getState(tgUserID int64) *userState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.states[tgUserID]
}

// --- Message helpers ---

// sendText sends a plain text message.
func (b *Bot) sendText(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		b.log.Error("failed to send message", "error", err)
	}
}

// sendTextWithKeyboard sends a text message with an inline keyboard.
func (b *Bot) sendTextWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = &keyboard
	if _, err := b.api.Send(msg); err != nil {
		b.log.Error("failed to send message", "error", err)
	}
}

// editMessageText edits an existing message's text and inline keyboard.
func (b *Bot) editMessageText(chatID int64, msgID int, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	edit := tgbotapi.NewEditMessageText(chatID, msgID, text)
	edit.ReplyMarkup = &keyboard
	if _, err := b.api.Send(edit); err != nil {
		b.log.Error("failed to edit message", "error", err)
	}
}
