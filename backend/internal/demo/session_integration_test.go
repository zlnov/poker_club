package demo

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"poker-club/backend/internal/config"
	"poker-club/backend/internal/repository/postgres"
)

// Requires a local poker_club_demo with the Asgard template already seeded.
func TestCreateSession_CloneAgainstDemoDB(t *testing.T) {
	if os.Getenv("DEMO_INTEGRATION") != "1" {
		t.Skip("set DEMO_INTEGRATION=1 to run")
	}

	cfg := config.Load()
	cfg.DemoEnabled = true
	if cfg.DemoDBUser == "" || cfg.DemoDBUser == "postgres" {
		if u := os.Getenv("DB_USER"); u != "" {
			cfg.DemoDBUser = u
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := postgres.NewDemo(ctx, cfg)
	if err != nil {
		t.Fatalf("connect demo db: %v", err)
	}
	defer db.Close()

	repos := postgres.NewRepositories(db)
	svc := NewService(db, repos, slog.Default())

	result, err := svc.CreateSession(ctx)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if result.AccessToken == "" || result.ClubID == 0 || result.PlayerID == 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
	t.Logf("session ok club_id=%d player_id=%d", result.ClubID, result.PlayerID)
}
