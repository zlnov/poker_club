package persistence

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Config holds database and JWT configuration
type Config struct {
	Host             string
	Port             string
	User             string
	Password         string
	DBName           string
	SSLMode          string
	JWTAccessSecret  string
	JWTRefreshSecret string
}

// LoadEnv loads environment variables
func LoadEnv() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Host:             getEnv("DB_HOST", "localhost"),
		Port:             getEnv("DB_PORT", "5432"),
		User:             getEnv("DB_USER", "postgres"),
		Password:         getEnv("DB_PASSWORD", "password"),
		DBName:           getEnv("DB_NAME", "poker_club"),
		SSLMode:          getEnv("DB_SSLMODE", "disable"),
		JWTAccessSecret:  getEnv("JWT_ACCESS_SECRET", "your-access-secret-key-change-in-production"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-key-change-in-production"),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Connect establishes connection to the database
func Connect(cfg *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	// Simple logger configuration
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword compares a hashed password with a plain password
func CheckPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
