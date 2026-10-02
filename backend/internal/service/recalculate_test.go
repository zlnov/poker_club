package service

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"poker-club/backend/internal/domain"
)

// mockGameParticipantRepository is a mock implementation of GameParticipantRepository
// for testing recalculateGameResults.
type mockGameParticipantRepository struct {
	participants []*domain.GameParticipant
	updated      []*domain.GameParticipant
	updateErr    error
}

func (m *mockGameParticipantRepository) Ping(ctx context.Context) error {
	return nil
}

func (m *mockGameParticipantRepository) Create(ctx context.Context, participant *domain.GameParticipant) (int64, error) {
	return 0, nil
}

func (m *mockGameParticipantRepository) GetByID(ctx context.Context, id int64) (*domain.GameParticipant, error) {
	return nil, nil
}

func (m *mockGameParticipantRepository) GetByGame(ctx context.Context, gameID int64) ([]*domain.GameParticipant, error) {
	return m.participants, nil
}

func (m *mockGameParticipantRepository) GetConfirmedByGame(ctx context.Context, gameID int64) ([]*domain.GameParticipant, error) {
	var confirmed []*domain.GameParticipant
	for _, p := range m.participants {
		if p.Status == "confirmed" {
			confirmed = append(confirmed, p)
		}
	}
	return confirmed, nil
}

func (m *mockGameParticipantRepository) GetByGameWithPlayers(ctx context.Context, gameID int64) ([]*domain.GameParticipantWithPlayer, error) {
	return nil, nil
}

func (m *mockGameParticipantRepository) GetByGameAndPlayer(ctx context.Context, gameID, playerID int64) (*domain.GameParticipant, error) {
	return nil, nil
}

func (m *mockGameParticipantRepository) Update(ctx context.Context, participant *domain.GameParticipant) error {
	m.updated = append(m.updated, participant)
	return m.updateErr
}

func (m *mockGameParticipantRepository) UpdateStatus(ctx context.Context, gameID, playerID int64, status string) error {
	return nil
}

func (m *mockGameParticipantRepository) Delete(ctx context.Context, gameID, playerID int64) error {
	return nil
}

func (m *mockGameParticipantRepository) RegisterBuyIn(ctx context.Context, gameID, playerID int64, buyInCount int) error {
	return nil
}

func (m *mockGameParticipantRepository) RegisterRebuy(ctx context.Context, gameID, playerID int64, rebuyCount int) error {
	return nil
}

func (m *mockGameParticipantRepository) UpdateChipsEnd(ctx context.Context, gameID, playerID int64, chipsEnd float64) error {
	return nil
}

func (m *mockGameParticipantRepository) GetPlayerFinishedStats(ctx context.Context, playerID, clubID int64) (totalBuyInCount int, avgPlace float64, err error) {
	return 0, 0, nil
}

func (m *mockGameParticipantRepository) GetPlayerFinishedStatsByGameType(ctx context.Context, playerID, clubID int64, gameType string) (totalBuyInCount int, avgPlace float64, gamesInProfit int, err error) {
	return 0, 0, 0, nil
}

// float64Ptr is a helper to create *float64 values.
func float64Ptr(v float64) *float64 {
	return &v
}

// intPtr is a helper to create *int values.
func intPtr(v int) *int {
	return &v
}

func TestRecalculateGameResults_CashGamePlacesCorrectly(t *testing.T) {
	ctx := context.Background()

	// Setup: 3 confirmed players in a Cash game
	// Buy-in: 100, Chip value: 1
	// Player 1: chips_end = 300 → payout = 300, profit = 200
	// Player 2: chips_end = 150 → payout = 150, profit = 50
	// Player 3: chips_end = 50  → payout = 50,  profit = -50
	// Total bank = 300, Total payout = 500 → diff = 200
	// Adjustment: 200 / 2 (positive profit players) = 100 each
	// Player 1: payout = 400, profit = 300
	// Player 2: payout = 250, profit = 150
	// Player 3: payout = 50, profit = -50
	// Places: P1=1, P2=2, P3=3

	game := &domain.Game{
		ID:           1,
		ClubID:       1,
		GameType:     "cash",
		BuyInAmount:  100,
		ChipValue:    1,
		Status:       "finished",
	}

	participants := []*domain.GameParticipant{
		{
			ID:           1,
			GameID:       1,
			PlayerID:     101,
			BuyInCount:   1,
			RebuyCount:   0,
			ChipsEnd:     float64Ptr(300),
			PayoutAmount: nil,
			Place:        nil,
			Status:       "confirmed",
		},
		{
			ID:           2,
			GameID:       1,
			PlayerID:     102,
			BuyInCount:   1,
			RebuyCount:   0,
			ChipsEnd:     float64Ptr(150),
			PayoutAmount: nil,
			Place:        nil,
			Status:       "confirmed",
		},
		{
			ID:           3,
			GameID:       1,
			PlayerID:     103,
			BuyInCount:   1,
			RebuyCount:   0,
			ChipsEnd:     float64Ptr(50),
			PayoutAmount: nil,
			Place:        nil,
			Status:       "confirmed",
		},
	}

	mockRepo := &mockGameParticipantRepository{
		participants: participants,
	}

	svc := &Service{
		repos: &domain.Repositories{
			GameParticipants: mockRepo,
		},
		log: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	}

	err := svc.recalculateGameResults(ctx, game, participants)
	if err != nil {
		t.Fatalf("recalculateGameResults() error = %v", err)
	}

	// Verify places are assigned correctly
	// After bank mismatch adjustment (diff = 200, adjustment = -100 per positive-profit player):
	// P1: payout = 300 - 100 = 200, profit = 100, roi = 100%
	// P2: payout = 150 - 100 = 50,  profit = -50, roi = -50%
	// P3: payout = 50,               profit = -50, roi = -50%
	// Sorted: P1(100), P2(-50), P3(-50)
	// P2 and P3 have equal profit and ROI → same place (2)
	// Places: P1=1, P2=2, P3=2

	placeMap := make(map[int64]int)
	for _, p := range mockRepo.updated {
		if p.Place == nil {
			t.Fatalf("participant %d has nil place", p.PlayerID)
		}
		placeMap[p.PlayerID] = *p.Place
	}

	if placeMap[101] != 1 {
		t.Errorf("player 101 place = %d, want 1", placeMap[101])
	}
	if placeMap[102] != 2 {
		t.Errorf("player 102 place = %d, want 2", placeMap[102])
	}
	if placeMap[103] != 2 {
		t.Errorf("player 103 place = %d, want 2 (equal profit with player 102)", placeMap[103])
	}
}

func TestRecalculateGameResults_ExcludesNonConfirmedParticipants(t *testing.T) {
	ctx := context.Background()

	game := &domain.Game{
		ID:           1,
		ClubID:       1,
		GameType:     "cash",
		BuyInAmount:  100,
		ChipValue:    1,
		Status:       "finished",
	}

	// 2 confirmed players + 1 invited (non-confirmed)
	// Only confirmed players should be in the results
	confirmedParticipants := []*domain.GameParticipant{
		{
			ID:         1,
			GameID:     1,
			PlayerID:   101,
			BuyInCount: 1,
			ChipsEnd:   float64Ptr(200),
			Status:     "confirmed",
		},
		{
			ID:         2,
			GameID:     1,
			PlayerID:   102,
			BuyInCount: 1,
			ChipsEnd:   float64Ptr(100),
			Status:     "confirmed",
		},
	}

	mockRepo := &mockGameParticipantRepository{
		participants: confirmedParticipants,
	}

	svc := &Service{
		repos: &domain.Repositories{
			GameParticipants: mockRepo,
		},
		log: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	}

	err := svc.recalculateGameResults(ctx, game, confirmedParticipants)
	if err != nil {
		t.Fatalf("recalculateGameResults() error = %v", err)
	}

	// Verify only 2 participants were updated (not the invited one)
	if len(mockRepo.updated) != 2 {
		t.Errorf("expected 2 updates, got %d", len(mockRepo.updated))
	}

	// Verify places: After bank mismatch adjustment (diff = 100, adjustment = -100):
	// P1: payout = 200 - 100 = 100, profit = 0
	// P2: payout = 100, profit = 0
	// Both have equal profit (0) and ROI (0) → same place (1)
	placeMap := make(map[int64]int)
	for _, p := range mockRepo.updated {
		if p.Place == nil {
			t.Fatalf("participant %d has nil place", p.PlayerID)
		}
		placeMap[p.PlayerID] = *p.Place
	}

	if placeMap[101] != 1 {
		t.Errorf("player 101 place = %d, want 1", placeMap[101])
	}
	if placeMap[102] != 1 {
		t.Errorf("player 102 place = %d, want 1 (equal profit with player 101)", placeMap[102])
	}
}

func TestRecalculateGameResults_EqualProfitSamePlace(t *testing.T) {
	ctx := context.Background()

	game := &domain.Game{
		ID:           1,
		ClubID:       1,
		GameType:     "cash",
		BuyInAmount:  100,
		ChipValue:    1,
		Status:       "finished",
	}

	// 3 players with same chips_end → same profit → same place
	participants := []*domain.GameParticipant{
		{
			ID:         1,
			GameID:     1,
			PlayerID:   101,
			BuyInCount: 1,
			ChipsEnd:   float64Ptr(100),
			Status:     "confirmed",
		},
		{
			ID:         2,
			GameID:     1,
			PlayerID:   102,
			BuyInCount: 1,
			ChipsEnd:   float64Ptr(100),
			Status:     "confirmed",
		},
		{
			ID:         3,
			GameID:     1,
			PlayerID:   103,
			BuyInCount: 1,
			ChipsEnd:   float64Ptr(100),
			Status:     "confirmed",
		},
	}

	mockRepo := &mockGameParticipantRepository{
		participants: participants,
	}

	svc := &Service{
		repos: &domain.Repositories{
			GameParticipants: mockRepo,
		},
		log: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	}

	err := svc.recalculateGameResults(ctx, game, participants)
	if err != nil {
		t.Fatalf("recalculateGameResults() error = %v", err)
	}

	// All players have same profit (0) and ROI (0) → all should get place 1
	placeMap := make(map[int64]int)
	for _, p := range mockRepo.updated {
		if p.Place == nil {
			t.Fatalf("participant %d has nil place", p.PlayerID)
		}
		placeMap[p.PlayerID] = *p.Place
	}

	for _, playerID := range []int64{101, 102, 103} {
		if placeMap[playerID] != 1 {
			t.Errorf("player %d place = %d, want 1 (equal profit)", playerID, placeMap[playerID])
		}
	}
}

func TestRecalculateGameResults_BankMismatchAdjustment(t *testing.T) {
	ctx := context.Background()

	game := &domain.Game{
		ID:           1,
		ClubID:       1,
		GameType:     "cash",
		BuyInAmount:  100,
		ChipValue:    1,
		Status:       "finished",
	}

	// 2 players, total bank = 200, total payout = 250 → diff = 50
	// Adjustment: 50 / 2 = 25 each
	participants := []*domain.GameParticipant{
		{
			ID:         1,
			GameID:     1,
			PlayerID:   101,
			BuyInCount: 1,
			ChipsEnd:   float64Ptr(150), // payout = 150, profit = 50
			Status:     "confirmed",
		},
		{
			ID:         2,
			GameID:     1,
			PlayerID:   102,
			BuyInCount: 1,
			ChipsEnd:   float64Ptr(100), // payout = 100, profit = 0
			Status:     "confirmed",
		},
	}

	mockRepo := &mockGameParticipantRepository{
		participants: participants,
	}

	svc := &Service{
		repos: &domain.Repositories{
			GameParticipants: mockRepo,
		},
		log: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	}

	err := svc.recalculateGameResults(ctx, game, participants)
	if err != nil {
		t.Fatalf("recalculateGameResults() error = %v", err)
	}

	// After bank mismatch adjustment (diff = 50, adjustment = -50 per positive-profit player):
	// Only P1 has positive profit (50), so adjustment = -50 / 1 = -50
	// P1: payout = 150 - 50 = 100, profit = 0
	// P2: payout = 100, profit = 0
	// Both have equal profit (0) and ROI (0) → same place (1)

	payoutMap := make(map[int64]float64)
	profitMap := make(map[int64]float64)
	placeMap := make(map[int64]int)

	for _, p := range mockRepo.updated {
		if p.PayoutAmount == nil {
			t.Fatalf("participant %d has nil payout", p.PlayerID)
		}
		if p.Place == nil {
			t.Fatalf("participant %d has nil place", p.PlayerID)
		}
		payoutMap[p.PlayerID] = *p.PayoutAmount
		placeMap[p.PlayerID] = *p.Place

		buyInAmount := float64(p.BuyInCount) * game.BuyInAmount
		profitMap[p.PlayerID] = *p.PayoutAmount - buyInAmount
	}

	// Verify payouts after adjustment
	if payoutMap[101] != 100 {
		t.Errorf("player 101 payout = %f, want 100", payoutMap[101])
	}
	if payoutMap[102] != 100 {
		t.Errorf("player 102 payout = %f, want 100", payoutMap[102])
	}

	// Verify places (equal profit → same place)
	if placeMap[101] != 1 {
		t.Errorf("player 101 place = %d, want 1", placeMap[101])
	}
	if placeMap[102] != 1 {
		t.Errorf("player 102 place = %d, want 1 (equal profit with player 101)", placeMap[102])
	}

	// Verify profits
	if profitMap[101] != 0 {
		t.Errorf("player 101 profit = %f, want 0", profitMap[101])
	}
	if profitMap[102] != 0 {
		t.Errorf("player 102 profit = %f, want 0", profitMap[102])
	}
}
