package usecases

import (
    "context"
    "errors"
    "poker-club-backend/domain"
)

// RegisterPlayerUseCase implements RegisterPlayerForGameUseCase.
// For now it simply validates that the player exists and returns nil.
// In a full implementation this would add the player to a game.
type RegisterPlayerUseCase struct {
	playerRepo domain.PlayerRepository
}

// NewRegisterPlayerUseCase creates a new instance.
func NewRegisterPlayerUseCase(playerRepo domain.PlayerRepository) *RegisterPlayerUseCase {
	return &RegisterPlayerUseCase{playerRepo: playerRepo}
}

// RegisterPlayerForGame satisfies the interface.
func (r *RegisterPlayerUseCase) RegisterPlayerForGame(ctx context.Context, gameID int64, playerID int64) error {
    if playerID == 0 {
        return errors.New("invalid player id")
    }
    // Try to get player
    _, err := r.playerRepo.GetByID(ctx, playerID)
    if err == nil {
        // already exists
        return nil
    }
    // Create a minimal player record
    p := &domain.Player{ID: playerID, FirstName: "TelegramUser"}
    return r.playerRepo.Create(ctx, p)
}
