package demo

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"poker-club/backend/internal/domain"
	"poker-club/backend/internal/repository/postgres"
)

const (
	// SessionTTL is the fixed lifetime of a demo session from creation.
	SessionTTL = 24 * time.Hour
	// CleanupInterval is how often expired demo sessions are purged.
	CleanupInterval = 12 * time.Hour
	tokenBytes      = 32
)

//go:embed asgard_demo_seed.sql
var seedFS embed.FS

// Service manages demo sessions against the demo database.
// It reuses domain.Repositories backed by the demo pool and never sets a Telegram notifier.
type Service struct {
	db    *postgres.DB
	repos *domain.Repositories
	log   *slog.Logger
}

// NewService creates a demo Service. repos must be wired to the demo database.
func NewService(db *postgres.DB, repos *domain.Repositories, log *slog.Logger) *Service {
	return &Service{db: db, repos: repos, log: log}
}

// Repos returns the demo-backed repositories for HTTP handlers.
func (s *Service) Repos() *domain.Repositories {
	return s.repos
}

// EnsureTemplate loads the Asgard template club once if it is missing.
func (s *Service) EnsureTemplate(ctx context.Context) error {
	var exists bool
	if err := s.db.Pool.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM clubs WHERE is_template = TRUE)
	`).Scan(&exists); err != nil {
		return fmt.Errorf("check demo template: %w", err)
	}
	if exists {
		return nil
	}

	seedSQL, err := seedFS.ReadFile("asgard_demo_seed.sql")
	if err != nil {
		return fmt.Errorf("read demo seed: %w", err)
	}

	sqlDB, err := sql.Open("pgx", s.db.DSN())
	if err != nil {
		return fmt.Errorf("open demo sql db: %w", err)
	}
	defer sqlDB.Close()

	if _, err := sqlDB.ExecContext(ctx, string(seedSQL)); err != nil {
		// Concurrent EnsureTemplate may race; treat "already exists" as success.
		if strings.Contains(err.Error(), "DEMO_TEMPLATE_EXISTS") {
			return nil
		}
		return fmt.Errorf("apply demo template seed: %w", err)
	}

	s.log.Info("demo template seeded", "club", "Asgard (Demo)")
	return nil
}

// CreateSession clones the Asgard template into a personal club and issues an opaque access token.
func (s *Service) CreateSession(ctx context.Context) (*domain.DemoSessionResult, error) {
	result, err := s.createSession(ctx)
	if err != nil {
		s.log.Error("demo session create failed", "error", err)
	}
	return result, err
}

func (s *Service) createSession(ctx context.Context) (*domain.DemoSessionResult, error) {
	if err := s.EnsureTemplate(ctx); err != nil {
		return nil, err
	}

	rawToken, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate demo token: %w", err)
	}
	tokenHash := hashToken(rawToken)
	expiresAt := time.Now().UTC().Add(SessionTTL)

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin demo session transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var templateClubID int64
	err = tx.QueryRow(ctx, `
		SELECT id FROM clubs WHERE is_template = TRUE ORDER BY id ASC LIMIT 1
	`).Scan(&templateClubID)
	if err != nil {
		return nil, fmt.Errorf("resolve demo template club: %w", err)
	}

	cloned, err := s.cloneTemplateClub(ctx, tx, templateClubID)
	if err != nil {
		return nil, err
	}

	var sessionID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO demo_sessions (token_hash, club_id, player_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id
	`, tokenHash, cloned.ClubID, cloned.PlayerID, expiresAt).Scan(&sessionID)
	if err != nil {
		return nil, fmt.Errorf("insert demo session: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit demo session: %w", err)
	}

	s.log.Info("demo session created",
		"session_id", sessionID,
		"club_id", cloned.ClubID,
		"player_id", cloned.PlayerID,
		"template_club_id", templateClubID,
		"expires_at", expiresAt,
	)

	return &domain.DemoSessionResult{
		AccessToken: rawToken,
		ClubID:      cloned.ClubID,
		ExpiresAt:   expiresAt,
		PlayerID:    cloned.PlayerID,
		TgUserID:    cloned.TgUserID,
	}, nil
}

// Authenticate validates an opaque demo token and returns the session if still valid.
func (s *Service) Authenticate(ctx context.Context, rawToken string) (*domain.DemoSession, *domain.Player, error) {
	if rawToken == "" {
		return nil, nil, domain.ErrNotFound
	}
	session, err := s.repos.DemoSessions.GetByTokenHash(ctx, hashToken(rawToken))
	if err != nil {
		return nil, nil, err
	}
	if time.Now().UTC().After(session.ExpiresAt.UTC()) {
		return nil, nil, domain.ErrNotFound
	}
	player, err := s.repos.Players.GetByID(ctx, session.PlayerID)
	if err != nil {
		return nil, nil, err
	}
	return session, player, nil
}

// CleanupExpired deletes expired demo sessions and their personal club data.
// Template clubs (is_template = TRUE) are never deleted.
func (s *Service) CleanupExpired(ctx context.Context) (int, error) {
	sessions, err := s.repos.DemoSessions.ListExpired(ctx, time.Now().UTC())
	if err != nil {
		return 0, err
	}

	removed := 0
	for _, session := range sessions {
		if err := s.deleteSessionData(ctx, session); err != nil {
			s.log.Warn("failed to cleanup demo session",
				"session_id", session.ID,
				"club_id", session.ClubID,
				"error", err,
			)
			continue
		}
		removed++
	}
	if removed > 0 {
		s.log.Info("demo sessions cleaned up", "count", removed)
	}
	return removed, nil
}

// StartCleanupLoop runs CleanupExpired on CleanupInterval until ctx is cancelled.
func (s *Service) StartCleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	if _, err := s.CleanupExpired(ctx); err != nil {
		s.log.Warn("demo cleanup failed", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			s.log.Info("demo cleanup loop stopped")
			return
		case <-ticker.C:
			if _, err := s.CleanupExpired(ctx); err != nil {
				s.log.Warn("demo cleanup failed", "error", err)
			}
		}
	}
}

func (s *Service) deleteSessionData(ctx context.Context, session *domain.DemoSession) error {
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var isTemplate bool
	if err := tx.QueryRow(ctx, `SELECT is_template FROM clubs WHERE id = $1`, session.ClubID).Scan(&isTemplate); err != nil {
		return fmt.Errorf("load club template flag: %w", err)
	}
	if isTemplate {
		// Never wipe the shared template; only drop the session row.
		if _, err := tx.Exec(ctx, `DELETE FROM demo_sessions WHERE id = $1`, session.ID); err != nil {
			return fmt.Errorf("delete demo_session: %w", err)
		}
		return tx.Commit(ctx)
	}

	rows, err := tx.Query(ctx, `SELECT player_id FROM club_members WHERE club_id = $1`, session.ClubID)
	if err != nil {
		return err
	}
	var playerIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		playerIDs = append(playerIDs, id)
	}
	rows.Close()

	if _, err := tx.Exec(ctx, `DELETE FROM events WHERE game_id IN (SELECT id FROM games WHERE club_id = $1)`, session.ClubID); err != nil {
		return fmt.Errorf("delete events: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM game_participants WHERE game_id IN (SELECT id FROM games WHERE club_id = $1)`, session.ClubID); err != nil {
		return fmt.Errorf("delete game_participants: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM games WHERE club_id = $1`, session.ClubID); err != nil {
		return fmt.Errorf("delete games: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM demo_sessions WHERE id = $1`, session.ID); err != nil {
		return fmt.Errorf("delete demo_session: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM clubs WHERE id = $1 AND is_template = FALSE`, session.ClubID); err != nil {
		return fmt.Errorf("delete club: %w", err)
	}

	for _, playerID := range playerIDs {
		var remaining int
		if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM club_members WHERE player_id = $1`, playerID).Scan(&remaining); err != nil {
			return err
		}
		if remaining == 0 {
			if _, err := tx.Exec(ctx, `DELETE FROM players WHERE id = $1`, playerID); err != nil {
				return fmt.Errorf("delete player %d: %w", playerID, err)
			}
		}
	}

	return tx.Commit(ctx)
}

func generateToken() (string, error) {
	buf := make([]byte, tokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// EnsureSessionClub rejects cross-session club access.
func EnsureSessionClub(session *domain.DemoSession, clubID int64) error {
	if session == nil || session.ClubID != clubID {
		return errors.New("access denied: demo session does not own this club")
	}
	return nil
}
