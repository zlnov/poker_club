package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Authentication errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// ClubService handles club-related business logic
type ClubService struct {
	clubRepo     ClubRepository
	memberRepo   ClubMemberRepository
	playerRepo   PlayerRepository
	chatClubRepo ChatClubRepository
}

// NewClubService creates a new ClubService
func NewClubService(
	clubRepo ClubRepository,
	memberRepo ClubMemberRepository,
	playerRepo PlayerRepository,
	chatClubRepo ChatClubRepository,
) *ClubService {
	return &ClubService{
		clubRepo:     clubRepo,
		memberRepo:   memberRepo,
		playerRepo:   playerRepo,
		chatClubRepo: chatClubRepo,
	}
}

// CreateClub creates a new club
func (s *ClubService) CreateClub(ctx context.Context, name string, creatorID int64) (*Club, error) {
	// Validate creator exists
	creator, err := s.playerRepo.GetByID(ctx, creatorID)
	if err != nil {
		return nil, ErrPlayerNotFound
	}
	if creator == nil {
		return nil, ErrPlayerNotFound
	}

	// Create club
	club := &Club{
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.clubRepo.Create(ctx, club); err != nil {
		return nil, err
	}

	// Add creator as admin member
	member := &ClubMember{
		ClubID:    club.ID,
		PlayerID:  creatorID,
		Role:      string(RoleAdmin),
		Status:    string(StatusActive),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	return club, nil
}

// ApproveMember approves a pending member
func (s *ClubService) ApproveMember(ctx context.Context, clubID, memberID int64, approverID int64) error {
	// Check if approver is admin of the club
	approver, err := s.memberRepo.GetByClubAndPlayer(ctx, clubID, approverID)
	if err != nil {
		return ErrUnauthorized
	}
	if approver == nil || approver.Role != string(RoleAdmin) {
		return ErrUnauthorized
	}

	// Get member to approve
	member, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return ErrMemberNotFound
	}
	if member == nil {
		return ErrMemberNotFound
	}

	// Check member belongs to the same club
	if member.ClubID != clubID {
		return ErrMemberNotFound
	}

	// Approve member
	member.Status = string(StatusActive)
	member.UpdatedAt = time.Now()

	return s.memberRepo.Update(ctx, member)
}

// RejectMember rejects a pending member
func (s *ClubService) RejectMember(ctx context.Context, clubID, memberID int64, rejecterID int64) error {
	// Check if rejecter is admin of the club
	rejecter, err := s.memberRepo.GetByClubAndPlayer(ctx, clubID, rejecterID)
	if err != nil {
		return ErrUnauthorized
	}
	if rejecter == nil || rejecter.Role != string(RoleAdmin) {
		return ErrUnauthorized
	}

	// Get member to reject
	member, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return ErrMemberNotFound
	}
	if member == nil {
		return ErrMemberNotFound
	}

	// Check member belongs to the same club
	if member.ClubID != clubID {
		return ErrMemberNotFound
	}

	// Delete the member record (rejection)
	return s.memberRepo.Delete(ctx, memberID)
}

// GetClubMembers returns all active members of a club
func (s *ClubService) GetClubMembers(ctx context.Context, clubID int64) ([]*ClubMember, error) {
	members, err := s.memberRepo.GetByClubID(ctx, clubID)
	if err != nil {
		return nil, err
	}

	// Filter only active members
	activeMembers := make([]*ClubMember, 0, len(members))
	for _, member := range members {
		if member.Status == string(StatusActive) {
			activeMembers = append(activeMembers, member)
		}
	}

	return activeMembers, nil
}

// CheckMemberAccess checks if a player has access to a club
func (s *ClubService) CheckMemberAccess(ctx context.Context, clubID, playerID int64) (bool, error) {
	member, err := s.memberRepo.GetByClubAndPlayer(ctx, clubID, playerID)
	if err != nil {
		return false, err
	}
	if member == nil {
		return false, nil
	}
	return member.Status == string(StatusActive), nil
}

// GameService handles game-related business logic
type GameService struct {
	gameRepo        GameRepository
	participantRepo GameParticipantRepository
	eventRepo       EventRepository
	memberRepo      ClubMemberRepository
	playerRepo      PlayerRepository
	clubRepo        ClubRepository
}

// NewGameService creates a new GameService
func NewGameService(
	gameRepo GameRepository,
	participantRepo GameParticipantRepository,
	eventRepo EventRepository,
	memberRepo ClubMemberRepository,
	playerRepo PlayerRepository,
	clubRepo ClubRepository,
) *GameService {
	return &GameService{
		gameRepo:        gameRepo,
		participantRepo: participantRepo,
		eventRepo:       eventRepo,
		memberRepo:      memberRepo,
		playerRepo:      playerRepo,
		clubRepo:        clubRepo,
	}
}

// CreateGameInput represents input for creating a game
type CreateGameInput struct {
	ClubID             int64
	BankerID           int64
	Type               GameType
	MoneyModel         MoneyModel
	BuyInAmount        float64
	RebuyAllowed       bool
	RebuyAmount        *float64
	MaxRebuysPerPlayer *int
	Duration           *time.Duration
	StartTime          time.Time
	MinPlayers         int
	MaxPlayers         int
	RankingPrimary     string
	RankingSecondary   *string
}

// BuyInInput represents input for buy-in
type BuyInInput struct {
	GameID      int64
	PlayerID    int64
	PerformedBy int64
}

// SetChipsInput represents input for setting final chips
type SetChipsInput struct {
	GameID      int64
	PlayerID    int64
	Chips       float64
	PerformedBy int64
}

// FinishGameInput represents input for finishing a game
type FinishGameInput struct {
	GameID      int64
	PerformedBy int64
}

// CreateGame creates a new game with banker as first participant
func (s *GameService) CreateGame(ctx context.Context, input CreateGameInput) (*Game, error) {
	// Validate game type
	if err := input.Type.Validate(); err != nil {
		return nil, err
	}

	// Validate money model
	if err := input.MoneyModel.Validate(); err != nil {
		return nil, err
	}

	// Check if banker is an active member of the club
	bankerMember, err := s.memberRepo.GetByClubAndPlayer(ctx, input.ClubID, input.BankerID)
	if err != nil {
		return nil, err
	}
	if bankerMember == nil || bankerMember.Status != string(StatusActive) {
		return nil, ErrMemberNotFound
	}

	// Validate rebuy settings
	if input.RebuyAllowed {
		if input.RebuyAmount == nil || *input.RebuyAmount <= 0 {
			return nil, errors.New("rebuy amount must be set when rebuy is allowed")
		}
	}

	// Create game
	game := &Game{
		ClubID:             input.ClubID,
		BankerID:           input.BankerID,
		Type:               string(input.Type),
		MoneyModel:         string(input.MoneyModel),
		BuyInAmount:        input.BuyInAmount,
		RebuyAllowed:       input.RebuyAllowed,
		RebuyAmount:        input.RebuyAmount,
		MaxRebuysPerPlayer: input.MaxRebuysPerPlayer,
		Duration:           input.Duration,
		StartTime:          input.StartTime,
		MinPlayers:         input.MinPlayers,
		MaxPlayers:         input.MaxPlayers,
		RankingPrimary:     input.RankingPrimary,
		RankingSecondary:   input.RankingSecondary,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.gameRepo.Create(ctx, game); err != nil {
		return nil, err
	}

	// Add banker as first participant
	participant := &GameParticipant{
		GameID:     game.ID,
		PlayerID:   input.BankerID,
		BuyInCount: 1,
		RebuyCount: 0,
		ChipsEnd:   &input.BuyInAmount,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.participantRepo.Create(ctx, participant); err != nil {
		return nil, err
	}

	// Record buy-in event for banker
	event := &Event{
		GameID:    game.ID,
		PlayerID:  input.BankerID,
		Type:      string(EventBuyIn),
		Amount:    &input.BuyInAmount,
		Metadata:  map[string]interface{}{"initial": true},
		CreatedAt: time.Now(),
		CreatedBy: input.BankerID,
	}

	if err := s.eventRepo.Create(ctx, event); err != nil {
		return nil, err
	}

	return game, nil
}

// BuyIn handles a player buy-in (initial or rebuy)
func (s *GameService) BuyIn(ctx context.Context, input BuyInInput) error {
	// Get game
	game, err := s.gameRepo.GetByID(ctx, input.GameID)
	if err != nil {
		return ErrGameNotFound
	}
	if game == nil {
		return ErrGameNotFound
	}

	// Check if game has ended
	if game.EndTime != nil {
		return ErrGameAlreadyEnded
	}

	// Check if player is active member of the club
	member, err := s.memberRepo.GetByClubAndPlayer(ctx, game.ClubID, input.PlayerID)
	if err != nil {
		return err
	}
	if member == nil || member.Status != string(StatusActive) {
		return ErrMemberNotFound
	}

	// Get or create participant
	participant, err := s.participantRepo.GetByGameAndPlayer(ctx, input.GameID, input.PlayerID)
	if err != nil {
		return err
	}

	if participant == nil {
		// First buy-in for this player
		participant = &GameParticipant{
			GameID:     input.GameID,
			PlayerID:   input.PlayerID,
			BuyInCount: 1,
			RebuyCount: 0,
			ChipsEnd:   &game.BuyInAmount,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := s.participantRepo.Create(ctx, participant); err != nil {
			return err
		}
	} else {
		// Check if this is a rebuy
		if !game.RebuyAllowed {
			return ErrRebuyNotAllowed
		}

		// Check max rebuys
		if game.MaxRebuysPerPlayer != nil {
			totalRebuys := participant.RebuyCount + 1
			if totalRebuys > *game.MaxRebuysPerPlayer {
				return ErrMaxRebuysExceeded
			}
		}

		// Update participant
		participant.RebuyCount++
		participant.UpdatedAt = time.Now()

		// Add rebuy amount to chips
		if participant.ChipsEnd == nil {
			chips := game.BuyInAmount
			participant.ChipsEnd = &chips
		} else {
			if game.RebuyAmount != nil {
				*participant.ChipsEnd += *game.RebuyAmount
			} else {
				*participant.ChipsEnd += game.BuyInAmount
			}
		}

		if err := s.participantRepo.Update(ctx, participant); err != nil {
			return err
		}
	}

	// Record event
	eventType := string(EventBuyIn)
	if participant.RebuyCount > 0 {
		eventType = string(EventRebuy)
	}

	event := &Event{
		GameID:    input.GameID,
		PlayerID:  input.PlayerID,
		Type:      eventType,
		Amount:    &game.BuyInAmount,
		Metadata:  map[string]interface{}{"buy_in_count": participant.BuyInCount, "rebuy_count": participant.RebuyCount},
		CreatedAt: time.Now(),
		CreatedBy: input.PerformedBy,
	}

	return s.eventRepo.Create(ctx, event)
}

// SetChips sets the final chips for a participant
func (s *GameService) SetChips(ctx context.Context, input SetChipsInput) error {
	// Get game
	game, err := s.gameRepo.GetByID(ctx, input.GameID)
	if err != nil {
		return ErrGameNotFound
	}
	if game == nil {
		return ErrGameNotFound
	}

	// Check if game has ended
	if game.EndTime == nil {
		return ErrGameNotStarted
	}

	// Get participant
	participant, err := s.participantRepo.GetByGameAndPlayer(ctx, input.GameID, input.PlayerID)
	if err != nil {
		return ErrParticipantNotFound
	}
	if participant == nil {
		return ErrParticipantNotFound
	}

	// Validate chips amount (should not be negative)
	if input.Chips < 0 {
		return ErrInvalidChipsAmount
	}

	// Update chips
	participant.ChipsEnd = &input.Chips
	participant.UpdatedAt = time.Now()

	if err := s.participantRepo.Update(ctx, participant); err != nil {
		return err
	}

	// Record event
	event := &Event{
		GameID:    input.GameID,
		PlayerID:  input.PlayerID,
		Type:      string(EventChipsSet),
		Amount:    &input.Chips,
		Metadata:  map[string]interface{}{},
		CreatedAt: time.Now(),
		CreatedBy: input.PerformedBy,
	}

	return s.eventRepo.Create(ctx, event)
}

// FinishGame finishes a game and validates chip totals
func (s *GameService) FinishGame(ctx context.Context, input FinishGameInput) error {
	// Get game
	game, err := s.gameRepo.GetByID(ctx, input.GameID)
	if err != nil {
		return ErrGameNotFound
	}
	if game == nil {
		return ErrGameNotFound
	}

	// Check if game already ended
	if game.EndTime != nil {
		return ErrGameAlreadyEnded
	}

	// Lock all participants for update
	participants, err := s.participantRepo.LockForUpdate(ctx, input.GameID)
	if err != nil {
		return err
	}

	// Validate that all participants have chips set
	totalInvested := 0.0
	totalChipsEnd := 0.0

	for _, p := range participants {
		// Calculate invested amount: buy_in_count * buy_in_amount + rebuy_count * rebuy_amount
		invested := game.BuyInAmount * float64(p.BuyInCount)
		if game.RebuyAmount != nil {
			invested += *game.RebuyAmount * float64(p.RebuyCount)
		}
		totalInvested += invested

		if p.ChipsEnd != nil {
			totalChipsEnd += *p.ChipsEnd
		}
	}

	// Validate chips match (with small tolerance for floating point)
	tolerance := 0.01
	if diff := totalInvested - totalChipsEnd; diff < -tolerance || diff > tolerance {
		return fmt.Errorf("%w: invested=%.2f, chips_end=%.2f", ErrChipsMismatch, totalInvested, totalChipsEnd)
	}

	// Set game end time
	now := time.Now()
	game.EndTime = &now
	game.UpdatedAt = now

	if err := s.gameRepo.Update(ctx, game); err != nil {
		return err
	}

	// Record finish event
	event := &Event{
		GameID:    input.GameID,
		PlayerID:  0, // System event
		Type:      string(EventCorrection),
		Metadata:  map[string]interface{}{"action": "finish_game", "total_invested": totalInvested, "total_chips_end": totalChipsEnd},
		CreatedAt: now,
		CreatedBy: input.PerformedBy,
	}

	return s.eventRepo.Create(ctx, event)
}

// GetGame returns a game by ID
func (s *GameService) GetGame(ctx context.Context, gameID int64) (*Game, error) {
	return s.gameRepo.GetByID(ctx, gameID)
}

// GetGameParticipants returns all participants of a game
func (s *GameService) GetGameParticipants(ctx context.Context, gameID int64) ([]*GameParticipant, error) {
	return s.participantRepo.GetByGameID(ctx, gameID)
}

// GetPlayerGames returns all games a player participated in
func (s *GameService) GetPlayerGames(ctx context.Context, playerID int64) ([]*Game, error) {
	participants, err := s.participantRepo.GetByPlayerID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	gameIDs := make([]int64, 0, len(participants))
	for _, p := range participants {
		gameIDs = append(gameIDs, p.GameID)
	}

	games := make([]*Game, 0, len(gameIDs))
	for _, gameID := range gameIDs {
		game, err := s.gameRepo.GetByID(ctx, gameID)
		if err != nil {
			continue
		}
		if game != nil {
			games = append(games, game)
		}
	}

	return games, nil
}

// ListGamesInput represents input for listing games
type ListGamesInput struct {
	ClubID int64
	Status *string
	Limit  int
	Offset int
}

// ListGamesOutput represents output from listing games
type ListGamesOutput struct {
	Games  []*Game
	Total  int64
	Limit  int
	Offset int
}

// GetGameDetailsInput represents input for getting game details
type GetGameDetailsInput struct {
	GameID int64
}

// GetGameDetailsOutput represents output from getting game details
type GetGameDetailsOutput struct {
	Game         *Game
	Participants []*GameParticipant
}

// GetLeaderboardInput represents input for getting leaderboard
type GetLeaderboardInput struct {
	ClubID int64
	Metric string
	Period string
}

// LeaderboardEntry represents a player's statistics for leaderboard
type LeaderboardEntry struct {
	PlayerID   int64
	PlayerName string
	Profit     float64
	ROI        float64
	WinRate    float64
	GamesCount int
}

// GetLeaderboardOutput represents output from getting leaderboard
type GetLeaderboardOutput struct {
	Metric   string
	Period   string
	Entries  []LeaderboardEntry
	ClubID   int64
	ClubName string
}

// ListGames returns games for a club with optional filters and pagination
func (s *GameService) ListGames(ctx context.Context, input ListGamesInput) (*ListGamesOutput, error) {
	games, err := s.gameRepo.GetByClubIDWithFilters(ctx, input.ClubID, input.Status, input.Limit, input.Offset)
	if err != nil {
		return nil, err
	}

	total, err := s.gameRepo.CountByClubID(ctx, input.ClubID, input.Status)
	if err != nil {
		return nil, err
	}

	return &ListGamesOutput{
		Games:  games,
		Total:  total,
		Limit:  input.Limit,
		Offset: input.Offset,
	}, nil
}

// GetGameDetails returns game with all participants
func (s *GameService) GetGameDetails(ctx context.Context, input GetGameDetailsInput) (*GetGameDetailsOutput, error) {
	game, err := s.gameRepo.GetByID(ctx, input.GameID)
	if err != nil {
		return nil, err
	}
	if game == nil {
		return nil, ErrGameNotFound
	}

	participants, err := s.participantRepo.GetByGameID(ctx, input.GameID)
	if err != nil {
		return nil, err
	}

	return &GetGameDetailsOutput{
		Game:         game,
		Participants: participants,
	}, nil
}

// GetPlayerName returns a player's full name
func (s *GameService) GetPlayerName(ctx context.Context, playerID int64) (string, error) {
	player, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return "", err
	}
	if player == nil {
		return "", ErrPlayerNotFound
	}
	return player.FirstName + " " + player.LastName, nil
}

// GetLeaderboard calculates leaderboard based on game statistics
func (s *GameService) GetLeaderboard(ctx context.Context, input GetLeaderboardInput) (*GetLeaderboardOutput, error) {
	// Get all finished games for the club
	status := "finished"
	games, err := s.gameRepo.GetByClubIDWithFilters(ctx, input.ClubID, &status, 1000, 0)
	if err != nil {
		return nil, err
	}

	// Get all participants from these games
	playerStats := make(map[int64]*LeaderboardEntry)

	for _, game := range games {
		participants, err := s.participantRepo.GetByGameID(ctx, game.ID)
		if err != nil {
			continue
		}

		for _, p := range participants {
			// Get player name
			playerName, err := s.GetPlayerName(ctx, p.PlayerID)
			if err != nil {
				// Skip if player not found
				continue
			}

			// Calculate invested amount
			invested := game.BuyInAmount * float64(p.BuyInCount)
			if game.RebuyAmount != nil {
				invested += *game.RebuyAmount * float64(p.RebuyCount)
			}

			// Calculate profit/loss
			chipsEnd := 0.0
			if p.ChipsEnd != nil {
				chipsEnd = *p.ChipsEnd
			}
			profit := chipsEnd - invested

			// Calculate ROI: (profit / invested) * 100
			roi := 0.0
			if invested > 0 {
				roi = (profit / invested) * 100
			}

			// Update or create stats
			if stats, exists := playerStats[p.PlayerID]; exists {
				stats.Profit += profit
				stats.ROI = ((stats.ROI * float64(stats.GamesCount)) + roi) / float64(stats.GamesCount+1) // average ROI
				stats.GamesCount++
				// Recalculate winrate based on place
				if p.Place != nil {
					// In The Money: place <= 3 (for MVP, assuming top 3 are ITM)
					if *p.Place <= 3 {
						stats.WinRate = ((stats.WinRate * float64(stats.GamesCount-1)) + 100) / float64(stats.GamesCount)
					} else {
						stats.WinRate = (stats.WinRate * float64(stats.GamesCount-1)) / float64(stats.GamesCount)
					}
				}
			} else {
				winRate := 0.0
				if p.Place != nil && *p.Place <= 3 {
					winRate = 100.0
				}
				playerStats[p.PlayerID] = &LeaderboardEntry{
					PlayerID:   p.PlayerID,
					PlayerName: playerName,
					Profit:     profit,
					ROI:        roi,
					WinRate:    winRate,
					GamesCount: 1,
				}
			}
		}
	}

	// Convert map to slice
	entries := make([]LeaderboardEntry, 0, len(playerStats))
	for _, entry := range playerStats {
		entries = append(entries, *entry)
	}

	// Sort based on metric
	switch input.Metric {
	case "profit":
		// Sort by profit descending
		for i := 0; i < len(entries); i++ {
			for j := i + 1; j < len(entries); j++ {
				if entries[i].Profit < entries[j].Profit {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}
	case "roi":
		// Sort by ROI descending
		for i := 0; i < len(entries); i++ {
			for j := i + 1; j < len(entries); j++ {
				if entries[i].ROI < entries[j].ROI {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}
	case "winrate":
		// Sort by WinRate descending
		for i := 0; i < len(entries); i++ {
			for j := i + 1; j < len(entries); j++ {
				if entries[i].WinRate < entries[j].WinRate {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}
	}

	// Get club name
	club, err := s.clubRepo.GetByID(ctx, input.ClubID)
	if err != nil {
		return nil, err
	}
	clubName := ""
	if club != nil {
		clubName = club.Name
	}

	return &GetLeaderboardOutput{
		Metric:   input.Metric,
		Period:   input.Period,
		Entries:  entries,
		ClubID:   input.ClubID,
		ClubName: clubName,
	}, nil
}
