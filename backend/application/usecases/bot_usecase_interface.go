package usecases

import (
	"context"
	"poker-club-backend/application/dtos"
	"poker-club-backend/domain"
)

// BotUseCase defines the contract for bot related operations.
// It extends existing use cases with game flow specific actions.
type BotUseCase interface {
	// Existing methods
	CreateGame(ctx context.Context, req dtos.CreateGameRequest) (*dtos.CreateGameResponse, error)
	RegisterPlayerForGame(ctx context.Context, gameID int64, playerID int64) error
	// New methods for game flow
	ProcessGameMenu(ctx context.Context, gameID int64, action string) (*dtos.GameMenuResponse, error)
	ProcessRebuy(ctx context.Context, gameID int64, playerID int64) (*dtos.BuyInResponse, error)
	ProcessGameStats(ctx context.Context, gameID int64) (*dtos.GameStatsResponse, error)
	ProcessFinishGame(ctx context.Context, gameID int64, results map[int64]float64) (*dtos.FinishGameResponse, error)
	ProcessCancelGame(ctx context.Context, gameID int64) error
	ProcessJoinPlayer(ctx context.Context, gameID int64, playerID int64) error
	// New methods for club and player lookup
	GetClubByChatID(ctx context.Context, chatID int64) (*domain.Club, error)
	HasAdminRights(ctx context.Context, chatID int64, clubID int64) bool
	PlayerExists(ctx context.Context, playerID int64) bool
}
