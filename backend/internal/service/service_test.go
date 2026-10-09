package service

import (
	"context"
	"testing"

	"poker-club/backend/internal/domain"
)

type mockPlayerRepository struct {
	playerToReturn *domain.Player
	getErr         error

	createdPlayer *domain.Player
	createID      int64
	createErr     error
	updatedTg     *string
}

func (f *mockPlayerRepository) Ping(ctx context.Context) error {
	return nil
}

func (f *mockPlayerRepository) Create(ctx context.Context, player *domain.Player) (int64, error) {
	f.createdPlayer = player
	return f.createID, f.createErr
}

func (f *mockPlayerRepository) GetByID(ctx context.Context, id int64) (*domain.Player, error) {
	return nil, domain.ErrNotFound
}

func (f *mockPlayerRepository) GetByTgUserID(ctx context.Context, tgUserID int64) (*domain.Player, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.playerToReturn, nil
}

func (f *mockPlayerRepository) GetByEmail(ctx context.Context, email string) (*domain.Player, error) {
	return nil, domain.ErrNotFound
}

func (f *mockPlayerRepository) GetByNickname(ctx context.Context, nickname string) (*domain.Player, error) {
	return nil, domain.ErrNotFound
}

func (f *mockPlayerRepository) GetByTgUserName(ctx context.Context, tgUserName string) (*domain.Player, error) {
	return nil, domain.ErrNotFound
}

func (f *mockPlayerRepository) UpdateLastSeen(ctx context.Context, id int64) error {
	return nil
}

func (f *mockPlayerRepository) UpdateTgUserName(ctx context.Context, id int64, tgUserName *string) error {
	f.updatedTg = tgUserName
	return nil
}

func (f *mockPlayerRepository) UpdateProfile(ctx context.Context, id int64, firstName, lastName string, nickname, email, phoneNumber *string) error {
	return nil
}

func (f *mockPlayerRepository) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	return nil
}

func TestRegisterTelegramUser_NewUser(t *testing.T) {
	ctx := context.Background()

	repo := &mockPlayerRepository{
		playerToReturn: nil,
		getErr:         domain.ErrNotFound,
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

	if err != nil {
		t.Fatalf("RegisterTelegramUser() error = %v", err)
	}

	if player == nil {
		t.Fatal("RegisterTelegramUser() returned nil player")
	}

	if player.ID != 123 {
		t.Errorf("player.ID = %d, want 123", player.ID)
	}

	if player.TgUserID == nil {
		t.Fatal("player.TgUserID = nil, want 100500")
	}

	if *player.TgUserID != 100500 {
		t.Errorf("player.TgUserID = %d, want 100500", *player.TgUserID)
	}

	if player.FirstName != "Ali" {
		t.Errorf("player.FirstName = %q, want %q", player.FirstName, "Ali")
	}

	if player.LastName != "Ze" {
		t.Errorf("player.LastName = %q, want %q", player.LastName, "Ze")
	}

	if player.Nickname != nil {
		t.Errorf("player.Nickname = %v, want nil", player.Nickname)
	}

	if player.TgUserName == nil || *player.TgUserName != "ze" {
		t.Errorf("player.TgUserName = %v, want ze", player.TgUserName)
	}

	if player.PhoneNumber != nil {
		t.Errorf("player.PhoneNumber = %v, want nil", *player.PhoneNumber)
	}

	if player.Email != nil {
		t.Errorf("player.Email = %v, want nil", *player.Email)
	}

	if player.Password != nil {
		t.Errorf("player.Password = %v, want nil", player.Password)
	}

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

func TestRegisterTelegramUser_EmptyUsername(t *testing.T) {
	ctx := context.Background()

	repo := &mockPlayerRepository{
		getErr:   domain.ErrNotFound,
		createID: 1,
	}

	svc := &Service{
		repos: &domain.Repositories{
			Players: repo,
		},
	}

	player, err := svc.RegisterTelegramUser(ctx, 42, "Ali", "Ze", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if player.TgUserName != nil {
		t.Errorf("expected nil tg_user_name, got %v", player.TgUserName)
	}
	if player.Nickname != nil {
		t.Errorf("expected nil nickname, got %v", player.Nickname)
	}
}
