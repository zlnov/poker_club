package domain

import (
	"context"
)

// StatisticsService provides methods to calculate player and game statistics.
type StatisticsService struct {
	gameRepo        GameRepository
	participantRepo GameParticipantRepository
	playerRepo      PlayerRepository
}

// NewStatisticsService creates a new StatisticsService.
func NewStatisticsService(
	gameRepo GameRepository,
	participantRepo GameParticipantRepository,
	playerRepo PlayerRepository,
) *StatisticsService {
	return &StatisticsService{
		gameRepo:        gameRepo,
		participantRepo: participantRepo,
		playerRepo:      playerRepo,
	}
}

// PlayerStats represents statistics for a player.
type PlayerStats struct {
	PlayerID      int64
	TotalGames    int
	TotalInvested float64
	TotalChips    float64
	Profit        float64
	ROI           float64
	ITM           float64
	WinRate       float64
}

// GameStats represents aggregate statistics for a finished game.
type GameStats struct {
	GameID        int64
	TotalInvested float64
	TotalChips    float64
}

// CalculatePlayerStats computes statistics for a player based on their game history.
func (s *StatisticsService) CalculatePlayerStats(ctx context.Context, playerID int64) (*PlayerStats, error) {
	// Get all game participant records for this player
	participants, err := s.participantRepo.GetByPlayerID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	var totalGames int
	var totalInvested float64
	var totalChips float64
	var itmCount int // In The Money count (placed in prize positions)
	var winCount int // First place wins

	// We need to know the prize places for each game to determine ITM.
	// For simplicity, we assume top 3 are ITM if not specified.
	// In a real system, prize places would be stored per game.
	for _, p := range participants {
		// Get the game details for buy-in, rebuy, etc.
		game, err := s.gameRepo.GetByID(ctx, p.GameID)
		if err != nil {
			continue
		}
		totalGames++

		// Invested = buy-in * buy-in count + rebuy amount * rebuy count
		invested := game.BuyInAmount * float64(p.BuyInCount)
		if game.RebuyAmount != nil {
			invested += *game.RebuyAmount * float64(p.RebuyCount)
		}
		totalInvested += invested

		// Chips won (chips_end)
		if p.ChipsEnd != nil {
			totalChips += *p.ChipsEnd
		}

		// Place (lower is better, 1st is best)
		if p.Place != nil {
			if *p.Place == 1 {
				winCount++
			}
			// Determine if ITM: assume top 3 are ITM
			if *p.Place <= 3 {
				itmCount++
			}
		}
	}

	stats := &PlayerStats{
		PlayerID:      playerID,
		TotalGames:    totalGames,
		TotalInvested: totalInvested,
		TotalChips:    totalChips,
		Profit:        totalChips - totalInvested,
	}

	if totalInvested > 0 {
		stats.ROI = (stats.Profit / totalInvested) * 100
	}
	if totalGames > 0 {
		stats.ITM = float64(itmCount) / float64(totalGames) * 100
		stats.WinRate = float64(winCount) / float64(totalGames) * 100
	}

	return stats, nil
}

// CalculateGameStats computes aggregate statistics for a finished game.
func (s *StatisticsService) CalculateGameStats(ctx context.Context, gameID int64) (*GameStats, error) {
	// Get the game
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return nil, err
	}
	// Get all participants for this game
	participants, err := s.participantRepo.GetByGameID(ctx, gameID)
	if err != nil {
		return nil, err
	}

	var totalInvested float64
	var totalChips float64
	for _, p := range participants {
		invested := game.BuyInAmount * float64(p.BuyInCount)
		if game.RebuyAmount != nil {
			invested += *game.RebuyAmount * float64(p.RebuyCount)
		}
		totalInvested += invested
		if p.ChipsEnd != nil {
			totalChips += *p.ChipsEnd
		}
	}

	return &GameStats{
		GameID:        game.ID,
		TotalInvested: totalInvested,
		TotalChips:    totalChips,
	}, nil
}
