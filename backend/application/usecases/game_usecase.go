package usecases

import (
	"context"

	"poker-club-backend/application/dtos"
	"poker-club-backend/domain"
)

// GameUseCase handles game-related use cases
type GameUseCase struct {
	gameService *domain.GameService
}

// NewGameUseCase creates a new GameUseCase
func NewGameUseCase(gameService *domain.GameService) *GameUseCase {
	return &GameUseCase{
		gameService: gameService,
	}
}

// CreateGame creates a new game
func (uc *GameUseCase) CreateGame(ctx context.Context, req dtos.CreateGameRequest) (*dtos.CreateGameResponse, error) {
	game, err := uc.gameService.CreateGame(ctx, domain.CreateGameInput{
		ClubID:             req.ClubID,
		BankerID:           req.BankerID,
		Type:               req.Type,
		MoneyModel:         req.MoneyModel,
		BuyInAmount:        req.BuyInAmount,
		RebuyAllowed:       req.RebuyAllowed,
		RebuyAmount:        req.RebuyAmount,
		MaxRebuysPerPlayer: req.MaxRebuysPerPlayer,
		Duration:           req.Duration,
		StartTime:          req.StartTime,
		MinPlayers:         req.MinPlayers,
		MaxPlayers:         req.MaxPlayers,
		RankingPrimary:     req.RankingPrimary,
		RankingSecondary:   req.RankingSecondary,
	})
	if err != nil {
		return nil, err
	}

	return &dtos.CreateGameResponse{
		ID:        game.ID,
		ClubID:    game.ClubID,
		BankerID:  game.BankerID,
		Type:      game.Type,
		StartTime: game.StartTime,
	}, nil
}

// BuyIn handles player buy-in or rebuy
func (uc *GameUseCase) BuyIn(ctx context.Context, req dtos.BuyInRequest) (*dtos.BuyInResponse, error) {
	if err := uc.gameService.BuyIn(ctx, domain.BuyInInput{
		GameID:      req.GameID,
		PlayerID:    req.PlayerID,
		PerformedBy: req.PerformedBy,
	}); err != nil {
		return nil, err
	}

	// Get participant to return current state
	participants, err := uc.gameService.GetGameParticipants(ctx, req.GameID)
	if err != nil {
		return nil, err
	}

	for _, p := range participants {
		if p.PlayerID == req.PlayerID {
			chipsEnd := 0.0
			if p.ChipsEnd != nil {
				chipsEnd = *p.ChipsEnd
			}
			return &dtos.BuyInResponse{
				GameID:     req.GameID,
				PlayerID:   req.PlayerID,
				BuyInCount: p.BuyInCount,
				RebuyCount: p.RebuyCount,
				ChipsEnd:   chipsEnd,
			}, nil
		}
	}

	return nil, domain.ErrParticipantNotFound
}

// SetChips sets final chips for a participant
func (uc *GameUseCase) SetChips(ctx context.Context, req dtos.SetChipsRequest) error {
	return uc.gameService.SetChips(ctx, domain.SetChipsInput{
		GameID:      req.GameID,
		PlayerID:    req.PlayerID,
		Chips:       req.Chips,
		PerformedBy: req.PerformedBy,
	})
}

// FinishGame finishes a game
func (uc *GameUseCase) FinishGame(ctx context.Context, req dtos.FinishGameRequest) (*dtos.FinishGameResponse, error) {
	if err := uc.gameService.FinishGame(ctx, domain.FinishGameInput{
		GameID:      req.GameID,
		PerformedBy: req.PerformedBy,
	}); err != nil {
		return nil, err
	}

	// Get game to return end time
	game, err := uc.gameService.GetGame(ctx, req.GameID)
	if err != nil {
		return nil, err
	}

	// Calculate totals
	participants, err := uc.gameService.GetGameParticipants(ctx, req.GameID)
	if err != nil {
		return nil, err
	}

	totalInvested := 0.0
	totalChips := 0.0
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

	return &dtos.FinishGameResponse{
		GameID:        req.GameID,
		EndTime:       *game.EndTime,
		TotalInvested: totalInvested,
		TotalChips:    totalChips,
	}, nil
}

// GetGameParticipants returns all participants of a game
func (uc *GameUseCase) GetGameParticipants(ctx context.Context, gameID int64) ([]dtos.GameParticipantDTO, error) {
	participants, err := uc.gameService.GetGameParticipants(ctx, gameID)
	if err != nil {
		return nil, err
	}

	result := make([]dtos.GameParticipantDTO, 0, len(participants))
	for _, p := range participants {
		dto := dtos.GameParticipantDTO{
			ID:         p.ID,
			PlayerID:   p.PlayerID,
			BuyInCount: p.BuyInCount,
			RebuyCount: p.RebuyCount,
			ChipsEnd:   0.0,
			Place:      p.Place,
		}
		if p.ChipsEnd != nil {
			dto.ChipsEnd = *p.ChipsEnd
		}
		result = append(result, dto)
	}

	return result, nil
}

// GetPlayerStats returns statistics for a player
func (uc *GameUseCase) GetPlayerStats(ctx context.Context, playerID int64) (*dtos.GetPlayerStatsResponse, error) {
	games, err := uc.gameService.GetPlayerGames(ctx, playerID)
	if err != nil {
		return nil, err
	}

	stats := &dtos.GetPlayerStatsResponse{
		PlayerID: playerID,
	}

	// This is a simplified version - in production would query events directly
	// For MVP we'll calculate from games and participants
	for _, game := range games {
		participants, err := uc.gameService.GetGameParticipants(ctx, game.ID)
		if err != nil {
			continue
		}

		for _, p := range participants {
			if p.PlayerID == playerID {
				stats.TotalGames++
				invested := game.BuyInAmount * float64(p.BuyInCount)
				if game.RebuyAmount != nil {
					invested += *game.RebuyAmount * float64(p.RebuyCount)
				}
				stats.TotalInvested += invested
				if p.ChipsEnd != nil {
					stats.TotalChips += *p.ChipsEnd
				}
				break
			}
		}
	}

	// Calculate profit and ROI
	stats.Profit = stats.TotalChips - stats.TotalInvested
	if stats.TotalInvested > 0 {
		stats.ROI = (stats.Profit / stats.TotalInvested) * 100
	}

	// Calculate ITM (In The Money) - simplified for MVP
	// In real implementation would track places across all games
	stats.ITM = 0

	return stats, nil
}

// ListGames returns list of games for a club with filters
func (uc *GameUseCase) ListGames(ctx context.Context, req dtos.ListGamesRequest) (*dtos.ListGamesResponse, error) {
	// Convert status to pointer
	var statusPtr *string
	if req.Status != "" {
		statusPtr = &req.Status
	}

	output, err := uc.gameService.ListGames(ctx, domain.ListGamesInput{
		ClubID: req.ClubID,
		Status: statusPtr,
		Limit:  req.Limit,
		Offset: req.Offset,
	})
	if err != nil {
		return nil, err
	}

	// Convert games to DTO
	gameDTOs := make([]dtos.GameListItemDTO, 0, len(output.Games))
	for _, game := range output.Games {
		// Count participants
		participants, err := uc.gameService.GetGameParticipants(ctx, game.ID)
		if err != nil {
			// If error, still include game with 0 participants
			gameDTOs = append(gameDTOs, dtos.GameListItemDTO{
				ID:           game.ID,
				ClubID:       game.ClubID,
				Type:         game.Type,
				MoneyModel:   game.MoneyModel,
				BuyInAmount:  game.BuyInAmount,
				StartTime:    game.StartTime,
				EndTime:      game.EndTime,
				Participants: 0,
			})
			continue
		}

		gameDTOs = append(gameDTOs, dtos.GameListItemDTO{
			ID:           game.ID,
			ClubID:       game.ClubID,
			Type:         game.Type,
			MoneyModel:   game.MoneyModel,
			BuyInAmount:  game.BuyInAmount,
			StartTime:    game.StartTime,
			EndTime:      game.EndTime,
			Participants: len(participants),
		})
	}

	return &dtos.ListGamesResponse{
		Games:  gameDTOs,
		Total:  output.Total,
		Limit:  output.Limit,
		Offset: output.Offset,
	}, nil
}

// GetGameDetails returns game with all participants
func (uc *GameUseCase) GetGameDetails(ctx context.Context, gameID int64) (*dtos.GetGameDetailsResponse, error) {
	output, err := uc.gameService.GetGameDetails(ctx, domain.GetGameDetailsInput{
		GameID: gameID,
	})
	if err != nil {
		return nil, err
	}

	// Convert game to response DTO
	gameResp := dtos.ToGameResponse(output.Game)

	// Convert participants to detail DTOs
	participantDTOs := make([]dtos.GameParticipantDetailDTO, 0, len(output.Participants))
	for _, p := range output.Participants {
		// Get player name
		playerName, err := uc.gameService.GetPlayerName(ctx, p.PlayerID)
		if err != nil {
			// If player not found, use placeholder
			playerName = "Unknown"
		}

		chipsEnd := 0.0
		if p.ChipsEnd != nil {
			chipsEnd = *p.ChipsEnd
		}

		participantDTOs = append(participantDTOs, dtos.GameParticipantDetailDTO{
			ID:         p.ID,
			PlayerID:   p.PlayerID,
			PlayerName: playerName,
			BuyInCount: p.BuyInCount,
			RebuyCount: p.RebuyCount,
			ChipsEnd:   chipsEnd,
			Place:      p.Place,
		})
	}

	return &dtos.GetGameDetailsResponse{
		Game:         gameResp,
		Participants: participantDTOs,
	}, nil
}

// GetLeaderboard returns leaderboard for a club
func (uc *GameUseCase) GetLeaderboard(ctx context.Context, req dtos.GetLeaderboardRequest) (*dtos.GetLeaderboardResponse, error) {
	output, err := uc.gameService.GetLeaderboard(ctx, domain.GetLeaderboardInput{
		ClubID: req.ClubID,
		Metric: req.Metric,
		Period: req.Period,
	})
	if err != nil {
		return nil, err
	}

	// Convert entries to DTOs
	entryDTOs := make([]dtos.LeaderboardEntryDTO, 0, len(output.Entries))
	for _, entry := range output.Entries {
		entryDTOs = append(entryDTOs, dtos.LeaderboardEntryDTO{
			PlayerID:   entry.PlayerID,
			PlayerName: entry.PlayerName,
			Metric:     getMetricValue(entry, req.Metric),
			GamesCount: entry.GamesCount,
		})
	}

	return &dtos.GetLeaderboardResponse{
		Metric:   output.Metric,
		Period:   output.Period,
		Entries:  entryDTOs,
		ClubID:   output.ClubID,
		ClubName: output.ClubName,
	}, nil
}

// Helper function to get metric value based on metric type
func getMetricValue(entry domain.LeaderboardEntry, metric string) float64 {
	switch metric {
	case "profit":
		return entry.Profit
	case "roi":
		return entry.ROI
	case "winrate":
		return entry.WinRate
	default:
		return 0
	}
}
