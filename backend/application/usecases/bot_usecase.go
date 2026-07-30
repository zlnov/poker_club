package usecases

import (
	"context"
	"poker-club-backend/application/dtos"
	"poker-club-backend/domain"
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

// The following methods are placeholders to satisfy the BotUseCase interface.
// Full implementations will delegate to underlying use cases and perform
// necessary business logic. For now they return nil or empty responses.

func (b *BotUseCaseImpl) ProcessGameMenu(ctx context.Context, gameID int64, action string) (*dtos.GameMenuResponse, error) {
	return &dtos.GameMenuResponse{GameID: gameID, MenuType: action}, nil
}

func (b *BotUseCaseImpl) ProcessRebuy(ctx context.Context, gameID int64, playerID int64) (*dtos.BuyInResponse, error) {
	return &dtos.BuyInResponse{}, nil
}

func (b *BotUseCaseImpl) ProcessGameStats(ctx context.Context, gameID int64) (*dtos.GameStatsResponse, error) {
	return &dtos.GameStatsResponse{GameID: gameID}, nil
}

func (b *BotUseCaseImpl) ProcessFinishGame(ctx context.Context, gameID int64, results map[int64]float64) (*dtos.FinishGameResponse, error) {
	return &dtos.FinishGameResponse{}, nil
}

func (b *BotUseCaseImpl) ProcessCancelGame(ctx context.Context, gameID int64) error {
	return nil
}

func (b *BotUseCaseImpl) ProcessJoinPlayer(ctx context.Context, gameID int64, playerID int64) error {
	return b.RegisterUseCase.RegisterPlayerForGame(ctx, gameID, playerID)
}

// GetClubByChatID retrieves the club associated with a Telegram chat ID.
// It queries the club repository via the game service which holds a reference
// to the club repository. The service layer is used to keep the use‑case
// independent from persistence details.
func (b *BotUseCaseImpl) GetClubByChatID(ctx context.Context, chatID int64) (*domain.Club, error) {
	// The game service exposes a method to get a club by chat ID.
	return b.GameUseCase.GetClubByChatID(ctx, chatID)
}

// HasAdminRights checks whether the user identified by chatID is an admin of
// the specified club. It uses the club member repository to look up the
// membership record and verifies the role.
func (b *BotUseCaseImpl) HasAdminRights(ctx context.Context, chatID int64, clubID int64) bool {
	member, err := b.GameUseCase.GetClubMemberByChatID(ctx, chatID, clubID)
	if err != nil || member == nil {
		return false
	}
	return member.Role == "admin"
}

// PlayerExists checks if a player with the given ID exists in the system.
func (b *BotUseCaseImpl) PlayerExists(ctx context.Context, playerID int64) bool {
	_, err := b.GameUseCase.GetPlayerByID(ctx, playerID)
	return err == nil
}
