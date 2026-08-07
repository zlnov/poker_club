package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"poker-club/backend/internal/config"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// DB wraps a connection pool and provides repository access.
type DB struct {
	Pool *pgxpool.Pool
}

// New creates a new DB instance, runs migrations, and returns it.
func New(ctx context.Context, cfg *config.Config) (*DB, error) {
	dsn := buildDSN(cfg)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{Pool: pool}

	if err := db.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func buildDSN(cfg *config.Config) string {
	q := url.Values{}
	q.Set("sslmode", cfg.DBSSLMode)

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?%s",
		cfg.DBUser,
		url.QueryEscape(cfg.DBPassword),
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		q.Encode(),
	)
}

// runMigrations executes pending Goose migrations.
func (db *DB) runMigrations() error {
	sqlDB, err := sql.Open("pgx", db.DSN())
	if err != nil {
		return fmt.Errorf("failed to open sql db for migrations: %w", err)
	}
	defer sqlDB.Close()

	migrations, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to get migrations filesystem: %w", err)
	}

	provider, err := goose.NewProvider("postgres", sqlDB, migrations)
	if err != nil {
		return fmt.Errorf("failed to create goose provider: %w", err)
	}

	if _, err := provider.Up(context.Background()); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// DSN returns the connection string for the database.
func (db *DB) DSN() string {
	return db.Pool.Config().ConnString()
}

// Close closes the connection pool.
func (db *DB) Close() {
	db.Pool.Close()
}

// Ping checks the database connection.
func (db *DB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}
