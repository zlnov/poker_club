package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"poker-club-backend/application/dtos"
	"poker-club-backend/domain"
)

// CommandHandler processes text commands from Telegram.
func (b *Bot) CommandHandler(ctx context.Context, msg *Message) {
	if !msg.IsCommand() {
		return
	}
	cmd := strings.ToLower(msg.Command())
	switch cmd {
	case "start":
		// Register player using chat ID as player identifier (placeholder)
		if err := b.usecase.RegisterPlayerForGame(context.Background(), 0, msg.Chat.ID); err != nil {
			b.sendText(msg.Chat.ID, "Error registering player")
			return
		}
		b.sendText(msg.Chat.ID, "Player registered successfully")
	case "new_game":
		// Example: /new_game <type> <banker_id> <buy_in> <min> <max>
		args := msg.CommandArguments()
		if args == "" {
			b.sendText(msg.Chat.ID, "Usage: /new_game <type> <banker_id> <buy_in> <min> <max>")
			return
		}
		parts := strings.Split(args, " ")
		if len(parts) < 5 {
			b.sendText(msg.Chat.ID, "Insufficient arguments for /new_game")
			return
		}
		// Parse arguments
		gameType := parts[0]
		bankerID, _ := strconv.ParseInt(parts[1], 10, 64)
		buyIn, _ := strconv.ParseFloat(parts[2], 64)
		minPlayers, _ := strconv.Atoi(parts[3])
		maxPlayers, _ := strconv.Atoi(parts[4])
		req := dtos.CreateGameRequest{
			ClubID:         0,
			BankerID:       bankerID,
			Type:           domain.GameType(gameType),
			MoneyModel:     domain.MoneyModel("real"),
			BuyInAmount:    buyIn,
			StartTime:      time.Now(),
			MinPlayers:     minPlayers,
			MaxPlayers:     maxPlayers,
			RankingPrimary: "chips",
		}
		resp, err := b.usecase.CreateGame(ctx, req)
		if err != nil {
			b.sendText(msg.Chat.ID, "Error creating game: "+err.Error())
			return
		}
		b.sendText(msg.Chat.ID, fmt.Sprintf("Game created with ID %d", resp.ID))
	// New game flow commands
	case "game_menu":
		b.handleGameMenu(ctx, msg)
	case "rebuy":
		b.handleRebuy(ctx, msg)
	case "game_stats":
		b.handleGameStats(ctx, msg)
	case "finish_game":
		b.handleFinishGame(ctx, msg)
	case "cancel_game":
		b.handleCancelGame(ctx, msg)
	case "join_player":
		b.handleJoinPlayer(ctx, msg)
	default:
		b.sendText(msg.Chat.ID, "Unknown command")
	}
}

// Stub implementations for new commands – these will be fleshed out later.
func (b *Bot) handleGameMenu(ctx context.Context, msg *Message) {
    // TODO: implement game menu logic
    b.sendText(msg.Chat.ID, "Game menu not implemented yet")
}
func (b *Bot) handleRebuy(ctx context.Context, msg *Message) {
    // TODO: implement rebuy logic
    b.sendText(msg.Chat.ID, "Rebuy not implemented yet")
}
func (b *Bot) handleGameStats(ctx context.Context, msg *Message) {
    // TODO: implement game stats logic
    b.sendText(msg.Chat.ID, "Game stats not implemented yet")
}
func (b *Bot) handleFinishGame(ctx context.Context, msg *Message) {
    // TODO: implement finish game logic
    b.sendText(msg.Chat.ID, "Finish game not implemented yet")
}
func (b *Bot) handleCancelGame(ctx context.Context, msg *Message) {
    // TODO: implement cancel game logic
    b.sendText(msg.Chat.ID, "Cancel game not implemented yet")
}
func (b *Bot) handleJoinPlayer(ctx context.Context, msg *Message) {
    // TODO: implement join player logic
    b.sendText(msg.Chat.ID, "Join player not implemented yet")
}
    // No additional method definitions – they are in bot.go
