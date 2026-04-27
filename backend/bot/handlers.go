package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"poker-club-backend/application/dtos"
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
		// Example: /new_game <game_name>
		args := msg.CommandArguments()
		if args == "" {
			b.sendText(msg.Chat.ID, "Usage: /new_game <name>")
			return
		}
		// Call usecase to create game
		// Construct a minimal valid CreateGameRequest using placeholder values.
		// In a real implementation these would be parsed from the command arguments.
		req := dtos.CreateGameRequest{
			ClubID:         0,
			BankerID:       0,
			Type:           "", // placeholder
			MoneyModel:     "", // placeholder
			BuyInAmount:    0,
			StartTime:      time.Now(),
			MinPlayers:     2,
			MaxPlayers:     10,
			RankingPrimary: "chips",
		}
		resp, err := b.usecase.CreateGame(ctx, req)
		if err != nil {
			b.sendText(msg.Chat.ID, "Error creating game")
			return
		}
		b.sendText(msg.Chat.ID, fmt.Sprintf("Game created with ID %d", resp.ID))
	default:
		b.sendText(msg.Chat.ID, "Unknown command")
	}
}
