package bot

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"poker-club-backend/application/usecases"
)

// BotUseCase defines the minimal use‑case methods the bot needs.
// BotUseCase defines the minimal set of use‑case methods the bot requires.
// It is implemented by usecases.BotUseCaseImpl.
type BotUseCase interface {
	usecases.BotUseCase
}

// Bot represents a Telegram bot instance.
// Minimal stub types to avoid external telegram dependency
// BotAPI wraps the real Telegram Bot API client
type BotAPI struct {
	client *tgbotapi.BotAPI
}

func NewBotAPI(token string) (*BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &BotAPI{client: bot}, nil
}

// Send forwards a message to the Telegram API. The argument must implement
// the tgbotapi.Chattable interface, which is satisfied by all message
// configuration types such as tgbotapi.MessageConfig, tgbotapi.PhotoConfig, etc.
func (b *BotAPI) Send(msg tgbotapi.Chattable) (interface{}, error) {
	return b.client.Send(msg)
}

// AnswerCallbackQuery sends a callback query answer. The argument should be
// of type tgbotapi.AnswerCallbackQueryConfig.
func (b *BotAPI) AnswerCallbackQuery(cb interface{}) error {
	// Stub: in production this would call b.client.AnswerCallbackQuery.
	return nil
}

type Chat struct{ ID int64 }
type Message struct {
	Chat Chat
	Text string
}

func (m *Message) IsCommand() bool { return strings.HasPrefix(m.Text, "/") }
func (m *Message) Command() string {
	parts := strings.SplitN(m.Text, " ", 2)
	return strings.TrimPrefix(parts[0], "/")
}
func (m *Message) CommandArguments() string {
	parts := strings.SplitN(m.Text, " ", 2)
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

type CallbackQuery struct {
	ID      string
	Data    string
	Message *Message
}
type Update struct {
	Message       *Message
	CallbackQuery *CallbackQuery
}

type Bot struct {
	api         *BotAPI
	usecase     BotUseCase
	webhookPath string
}

// Placeholder handlers for new commands – they will be fleshed out later.
// func (b *Bot) handleGameMenu(ctx context.Context, msg *Message) {
// 	b.sendText(msg.Chat.ID, "Game menu not implemented yet")
// }
// func (b *Bot) handleRebuy(ctx context.Context, msg *Message) {
// 	b.sendText(msg.Chat.ID, "Rebuy not implemented yet")
// }
// func (b *Bot) handleGameStats(ctx context.Context, msg *Message) {
// 	b.sendText(msg.Chat.ID, "Game stats not implemented yet")
// }
// func (b *Bot) handleFinishGame(ctx context.Context, msg *Message) {
// 	b.sendText(msg.Chat.ID, "Finish game not implemented yet")
// }
// func (b *Bot) handleCancelGame(ctx context.Context, msg *Message) {
// 	b.sendText(msg.Chat.ID, "Cancel game not implemented yet")
// }
// func (b *Bot) handleJoinPlayer(ctx context.Context, msg *Message) {
// 	b.sendText(msg.Chat.ID, "Join player not implemented yet")
// }

// NewBot creates a new Bot instance.
func NewBot(usecase BotUseCase, webhookPath string) (*Bot, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, os.ErrInvalid
	}
	api, err := NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	_ = api // no debug flag in stub
	return &Bot{api: api, usecase: usecase, webhookPath: webhookPath}, nil
}

// ServeWebhook starts an HTTP server to receive Telegram updates.
func (b *Bot) ServeWebhook(addr string) error {
	log.Println("Registering webhook path:", b.webhookPath)
	http.HandleFunc(b.webhookPath, b.handleUpdate)
	return http.ListenAndServe(addr, nil)
}

// handleUpdate processes incoming Telegram updates.
func (b *Bot) handleUpdate(w http.ResponseWriter, r *http.Request) {
	var upd Update
	if err := json.NewDecoder(r.Body).Decode(&upd); err != nil {
		log.Printf("failed to decode update: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("UPD message: %+v, UPD CallbackQuery: %+v\n", upd.Message, upd.CallbackQuery)

	// Simple routing based on update type
	switch {
	case upd.Message != nil:
		b.handleMessage(upd.Message)
	case upd.CallbackQuery != nil:
		b.handleCallback(upd.CallbackQuery)
	default:
		log.Printf("unknown update type")
	}
	log.Println("WEBHOOK HIT:", r.URL.Path)
	w.WriteHeader(http.StatusOK)
}

func (b *Bot) handleMessage(msg *Message) {
	// Delegate to CommandHandler for extensibility
	b.CommandHandler(context.Background(), msg)
}

func (b *Bot) handleCallback(cb *CallbackQuery) {
	data := cb.Data
	parts := strings.Split(data, ":")
	if len(parts) < 1 {
		return
	}
	action := parts[0]
	switch action {
	case "join":
		// placeholder: join game logic
		b.sendText(cb.Message.Chat.ID, "Joining game...")
	}
	// answer callback to remove loading state (stub)
	_ = b.api.AnswerCallbackQuery(nil)
}

func (b *Bot) sendText(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("failed to send message: %v", err)
	}
}
