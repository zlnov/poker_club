package service

import (
	"context"
	"errors"
	"testing"

	"poker-club/backend/internal/domain"
)

var errPlayerNotFound = errors.New("player not found")

type mockPlayerRepository struct {
	playerToReturn *domain.Player
	getErr         error

	createdPlayer *domain.Player
	createID      int64
	createErr     error
}

func (f *mockPlayerRepository) Ping(ctx context.Context) error {
	return nil
}

func (f *mockPlayerRepository) Create(ctx context.Context, player *domain.Player) (int64, error) {
	f.createdPlayer = player
	return f.createID, f.createErr
}

func (f *mockPlayerRepository) GetByID(ctx context.Context, id int64) (*domain.Player, error) {
	return nil, nil
}

func (f *mockPlayerRepository) GetByTgUserID(ctx context.Context, tgUserID int64) (*domain.Player, error) {
	return f.playerToReturn, f.getErr
}

func (f *mockPlayerRepository) GetByNickname(ctx context.Context, nickname string) (*domain.Player, error) {
	return nil, nil
}

func (f *mockPlayerRepository) UpdateLastSeen(ctx context.Context, id int64) error {
	return nil
}

func TestRegisterTelegramUser_NewUser(t *testing.T) {
	ctx := context.Background()

	repo := &mockPlayerRepository{
		playerToReturn: nil,
		getErr:         errPlayerNotFound,
		createID:       123,
	}

	svc := &Service{
		repos: &domain.Repositories{
			Players: repo,
		},
	}

	player, err := svc.RegisterTelegramUser(
		ctx,
		100500,
		"Ali",
		"Ze",
		"ze",
	)

	// Check result.
	if err != nil {
		t.Fatalf("RegisterTelegramUser() error = %v", err)
	}

	if player == nil {
		t.Fatal("RegisterTelegramUser() returned nil player")
	}

	// Check generated player ID.
	if player.ID != 123 {
		t.Errorf("player.ID = %d, want 123", player.ID)
	}

	// Check Telegram user ID.
	if player.TgUserID == nil {
		t.Fatal("player.TgUserID = nil, want 100500")
	}

	if *player.TgUserID != 100500 {
		t.Errorf("player.TgUserID = %d, want 100500", *player.TgUserID)
	}

	// Check Telegram user data.
	if player.FirstName != "Ali" {
		t.Errorf("player.FirstName = %q, want %q", player.FirstName, "Ali")
	}

	if player.LastName != "Ze" {
		t.Errorf("player.LastName = %q, want %q", player.LastName, "Ze")
	}

	if player.Nickname != "ze" {
		t.Errorf("player.Nickname = %q, want %q", player.Nickname, "ze")
	}

	// Telegram does not provide phone number and email.
	if player.PhoneNumber != nil {
		t.Errorf("player.PhoneNumber = %v, want nil", *player.PhoneNumber)
	}

	if player.Email != nil {
		t.Errorf("player.Email = %v, want nil", *player.Email)
	}

	// Check what was passed to repository.Create().
	if repo.createdPlayer == nil {
		t.Fatal("repository.Create() was not called")
	}

	if repo.createdPlayer.PhoneNumber != nil {
		t.Errorf("created player PhoneNumber = %v, want nil", *repo.createdPlayer.PhoneNumber)
	}

	if repo.createdPlayer.Email != nil {
		t.Errorf("created player Email = %v, want nil", *repo.createdPlayer.Email)
	}

	if repo.createdPlayer.TgUserID == nil {
		t.Fatal("created player TgUserID = nil")
	}

	if *repo.createdPlayer.TgUserID != 100500 {
		t.Errorf(
			"created player TgUserID = %d, want 100500",
			*repo.createdPlayer.TgUserID,
		)
	}
}
