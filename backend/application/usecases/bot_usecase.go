package usecases

import (
	"context"
	"poker-club-backend/application/dtos"
)

// BotUseCaseImpl combines game and register usecases to satisfy bot.BotUseCase.
type BotUseCaseImpl struct {
	GameUseCase     *GameUseCase
	RegisterUseCase *RegisterPlayerUseCase
}

// CreateGame delegates to GameUseCase.
func (b *BotUseCaseImpl) CreateGame(ctx context.Context, req dtos.CreateGameRequest) (*dtos.CreateGameResponse, error) {
	return b.GameUseCase.CreateGame(ctx, req)
}

// RegisterPlayerForGame delegates to RegisterUseCase.
func (b *BotUseCaseImpl) RegisterPlayerForGame(ctx context.Context, gameID int64, playerID int64) error {
	return b.RegisterUseCase.RegisterPlayerForGame(ctx, gameID, playerID)
}
