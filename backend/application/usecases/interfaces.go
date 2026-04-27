package usecases

import (
	"context"

	"poker-club-backend/application/dtos"
)

// CreateGameUseCase defines the contract for creating a game.
type CreateGameUseCase interface {
	CreateGame(ctx context.Context, req dtos.CreateGameRequest) (*dtos.CreateGameResponse, error)
}

// RegisterPlayerForGameUseCase defines the contract for registering a player for a game.
type RegisterPlayerForGameUseCase interface {
	RegisterPlayerForGame(ctx context.Context, gameID int64, playerID int64) error
}

// BuyInUseCase defines the contract for buying in or rebuying.
type BuyInUseCase interface {
	BuyIn(ctx context.Context, req dtos.BuyInRequest) (*dtos.BuyInResponse, error)
}

// SetChipsUseCase defines the contract for setting final chips of a participant.
type SetChipsUseCase interface {
	SetChips(ctx context.Context, req dtos.SetChipsRequest) error
}

// FinishGameUseCase defines the contract for finishing a game.
type FinishGameUseCase interface {
	FinishGame(ctx context.Context, req dtos.FinishGameRequest) (*dtos.FinishGameResponse, error)
}

// CancelGameUseCase defines the contract for cancelling a game.
type CancelGameUseCase interface {
	CancelGame(ctx context.Context, gameID int64, performedBy int64) error
}

// GetPlayerStatsUseCase defines the contract for retrieving player statistics.
type GetPlayerStatsUseCase interface {
	GetPlayerStats(ctx context.Context, playerID int64) (*dtos.GetPlayerStatsResponse, error)
}

// GetLeaderboardUseCase defines the contract for retrieving leaderboard.
type GetLeaderboardUseCase interface {
	GetLeaderboard(ctx context.Context, req dtos.GetLeaderboardRequest) (*dtos.GetLeaderboardResponse, error)
}

// ListGamesUseCase defines the contract for listing games with filters.
type ListGamesUseCase interface {
	ListGames(ctx context.Context, req dtos.ListGamesRequest) (*dtos.ListGamesResponse, error)
}

// GetGameDetailsUseCase defines the contract for retrieving game details with participants.
type GetGameDetailsUseCase interface {
	GetGameDetails(ctx context.Context, gameID int64) (*dtos.GetGameDetailsResponse, error)
}
