package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"poker-club/backend/internal/domain"
	"poker-club/backend/internal/service"
)

// --- Mock repositories for GetClub handler test ---

type mockClubRepo struct {
	club    *domain.Club
	clubErr error
}

func (m *mockClubRepo) Ping(ctx context.Context) error { return nil }
func (m *mockClubRepo) Create(ctx context.Context, club *domain.Club) (int64, error) {
	return 0, nil
}
func (m *mockClubRepo) GetByID(ctx context.Context, id int64) (*domain.Club, error) {
	if m.clubErr != nil {
		return nil, m.clubErr
	}
	if m.club != nil && m.club.ID == id {
		return m.club, nil
	}
	return nil, errors.New("club not found")
}
func (m *mockClubRepo) GetByOwner(ctx context.Context, playerID int64) ([]*domain.Club, error) {
	return nil, nil
}
func (m *mockClubRepo) GetByPlayer(ctx context.Context, playerID int64) ([]*domain.Club, error) {
	return nil, nil
}
func (m *mockClubRepo) GetByTgChatID(ctx context.Context, tgChatID int64) (*domain.Club, error) {
	return nil, nil
}
func (m *mockClubRepo) UpdateName(ctx context.Context, id int64, name string) error {
	return nil
}
func (m *mockClubRepo) BindTgChatID(ctx context.Context, clubID, tgChatID int64) error {
	return nil
}
func (m *mockClubRepo) Delete(ctx context.Context, id int64) error { return nil }

type mockPlayerRepo struct {
	player    *domain.Player
	playerErr error
}

func (m *mockPlayerRepo) Ping(ctx context.Context) error { return nil }
func (m *mockPlayerRepo) Create(ctx context.Context, player *domain.Player) (int64, error) {
	return 0, nil
}
func (m *mockPlayerRepo) GetByID(ctx context.Context, id int64) (*domain.Player, error) {
	if m.player != nil && m.player.ID == id {
		return m.player, nil
	}
	return nil, errors.New("player not found")
}
func (m *mockPlayerRepo) GetByTgUserID(ctx context.Context, tgUserID int64) (*domain.Player, error) {
	if m.playerErr != nil {
		return nil, m.playerErr
	}
	if m.player != nil && m.player.TgUserID != nil && *m.player.TgUserID == tgUserID {
		return m.player, nil
	}
	return nil, errors.New("player not found")
}
func (m *mockPlayerRepo) GetByNickname(ctx context.Context, nickname string) (*domain.Player, error) {
	return nil, nil
}
func (m *mockPlayerRepo) UpdateLastSeen(ctx context.Context, id int64) error { return nil }

type mockClubMemberRepo struct {
	member    *domain.ClubMember
	memberErr error
}

func (m *mockClubMemberRepo) Ping(ctx context.Context) error { return nil }
func (m *mockClubMemberRepo) Create(ctx context.Context, member *domain.ClubMember) (int64, error) {
	return 0, nil
}
func (m *mockClubMemberRepo) GetByClubAndPlayer(ctx context.Context, clubID, playerID int64) (*domain.ClubMember, error) {
	if m.memberErr != nil {
		return nil, m.memberErr
	}
	if m.member != nil && m.member.ClubID == clubID && m.member.PlayerID == playerID {
		return m.member, nil
	}
	return nil, errors.New("club member not found")
}
func (m *mockClubMemberRepo) GetByClubWithPlayers(ctx context.Context, clubID int64) ([]*domain.ClubMemberWithPlayer, error) {
	return nil, nil
}
func (m *mockClubMemberRepo) CountActiveMembers(ctx context.Context, clubID int64) (int, error) {
	return 0, nil
}
func (m *mockClubMemberRepo) UpdateRole(ctx context.Context, clubID, playerID int64, role string) error {
	return nil
}
func (m *mockClubMemberRepo) UpdateStatus(ctx context.Context, clubID, playerID int64, status string) error {
	return nil
}
func (m *mockClubMemberRepo) UpdateAccepted(ctx context.Context, clubID, playerID int64, accepted bool) error {
	return nil
}

// --- Test helpers ---

func setupTestGetClub(t *testing.T, player *domain.Player, member *domain.ClubMember, club *domain.Club) (*ClubHandler, *gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	playerRepo := &mockPlayerRepo{player: player}
	clubRepo := &mockClubRepo{club: club}
	memberRepo := &mockClubMemberRepo{member: member}

	repos := &domain.Repositories{
		Clubs:       clubRepo,
		Players:     playerRepo,
		ClubMembers: memberRepo,
	}

	logger := testLogger()
	svc := service.New(repos, logger)
	handler := NewClubHandler(svc)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/clubs/1", nil)
	c.Params = gin.Params{{Key: "clubId", Value: "1"}}

	// Set authenticated user in context
	authUser := &domain.AuthenticatedUser{
		PlayerID: player.ID,
		TgUserID: player.TgUserID,
		Role:     "member",
	}
	c.Set(string(authenticatedUserKey), authUser)

	return handler, c, w
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// --- Tests ---

func TestGetClub_AdminRole(t *testing.T) {
	tgUserID := int64(12345)
	player := &domain.Player{
		ID:        1,
		TgUserID:  &tgUserID,
		FirstName: "Test",
		LastName:  "User",
		Nickname:  "testuser",
	}
	club := &domain.Club{
		ID:        1,
		Name:      "Test Club",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	member := &domain.ClubMember{
		ID:        1,
		ClubID:    1,
		PlayerID:  1,
		Role:      "admin",
		Status:    "active",
		Accepted:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	handler, c, w := setupTestGetClub(t, player, member, club)
	handler.GetClub(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	isOwner, _ := response["is_owner"].(bool)
	isAdmin, _ := response["is_admin"].(bool)
	role, _ := response["role"].(string)

	if isOwner {
		t.Error("expected is_owner=false for admin role, got true")
	}
	if !isAdmin {
		t.Error("expected is_admin=true for admin role, got false")
	}
	if role != "admin" {
		t.Errorf("expected role='admin', got %q", role)
	}
}

func TestGetClub_OwnerRole(t *testing.T) {
	tgUserID := int64(12345)
	player := &domain.Player{
		ID:        1,
		TgUserID:  &tgUserID,
		FirstName: "Test",
		LastName:  "User",
		Nickname:  "testuser",
	}
	club := &domain.Club{
		ID:        1,
		Name:      "Test Club",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	member := &domain.ClubMember{
		ID:        1,
		ClubID:    1,
		PlayerID:  1,
		Role:      "owner",
		Status:    "active",
		Accepted:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	handler, c, w := setupTestGetClub(t, player, member, club)
	handler.GetClub(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	isOwner, _ := response["is_owner"].(bool)
	isAdmin, _ := response["is_admin"].(bool)
	role, _ := response["role"].(string)

	if !isOwner {
		t.Error("expected is_owner=true for owner role, got false")
	}
	if isAdmin {
		t.Error("expected is_admin=false for owner role, got true")
	}
	if role != "owner" {
		t.Errorf("expected role='owner', got %q", role)
	}
}

func TestGetClub_MemberRole(t *testing.T) {
	// Note: 'member' role does not have PermViewClub permission,
	// so GetClub will return 403 for members. This test verifies
	// that the handler correctly denies access for members.
	tgUserID := int64(12345)
	player := &domain.Player{
		ID:        1,
		TgUserID:  &tgUserID,
		FirstName: "Test",
		LastName:  "User",
		Nickname:  "testuser",
	}
	club := &domain.Club{
		ID:        1,
		Name:      "Test Club",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	member := &domain.ClubMember{
		ID:        1,
		ClubID:    1,
		PlayerID:  1,
		Role:      "member",
		Status:    "active",
		Accepted:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	handler, c, w := setupTestGetClub(t, player, member, club)
	handler.GetClub(c)

	// Member role does not have PermViewClub, so expect 403
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 for member role, got %d", w.Code)
	}
}

func TestGetClub_TgUserIDNotEqualPlayerID(t *testing.T) {
	// This test verifies the fix: when tgUserID != playerID,
	// the handler should resolve tgUserID to playerID before
	// looking up club membership.
	tgUserID := int64(99999) // Different from player.ID
	player := &domain.Player{
		ID:        1, // playerID is 1, but tgUserID is 99999
		TgUserID:  &tgUserID,
		FirstName: "Test",
		LastName:  "User",
		Nickname:  "testuser",
	}
	club := &domain.Club{
		ID:        1,
		Name:      "Test Club",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	member := &domain.ClubMember{
		ID:        1,
		ClubID:    1,
		PlayerID:  1, // playerID is 1, matching player.ID
		Role:      "admin",
		Status:    "active",
		Accepted:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	handler, c, w := setupTestGetClub(t, player, member, club)
	handler.GetClub(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	isOwner, _ := response["is_owner"].(bool)
	isAdmin, _ := response["is_admin"].(bool)
	role, _ := response["role"].(string)

	// Before the fix, is_admin would be false because GetClubMember
	// was called with tgUserID (99999) instead of playerID (1),
	// causing the member lookup to fail.
	if !isAdmin {
		t.Error("expected is_admin=true for admin role, got false — GetClubMember was likely called with tgUserID instead of playerID")
	}
	if isOwner {
		t.Error("expected is_owner=false for admin role, got true")
	}
	if role != "admin" {
		t.Errorf("expected role='admin', got %q", role)
	}
}

// Ensure testLogger is available
var _ = os.Stderr
