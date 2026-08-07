package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	BotToken    string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBSSLMode   string
	WebhookURL  string
	WebhookPort string
	LongPolling bool
	LogLevel    string
}

// Load reads configuration from environment variables and returns a Config.
func Load() *Config {
	return &Config{
		BotToken:    os.Getenv("BOT_TOKEN"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  os.Getenv("DB_PASSWORD"),
		DBName:      getEnv("DB_NAME", "poker_club"),
		DBSSLMode:   getEnv("DB_SSLMODE", "disable"),
		WebhookURL:  os.Getenv("WEBHOOK_URL"),
		WebhookPort: getEnv("WEBHOOK_PORT", "8080"),
		LongPolling: getEnvBool("LONG_POLLING", true),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

// DSN returns the PostgreSQL connection string.
func (c *Config) DSN() string {
	return "host=" + c.DBHost +
		" port=" + c.DBPort +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" sslmode=" + c.DBSSLMode +
		" pool_max_conns=10"
}

// WebhookPortDuration returns the webhook port with a colon prefix for Listen.
func (c *Config) WebhookAddr() string {
	return ":" + c.WebhookPort
}

// PollTimeout returns the long polling timeout duration.
func (c *Config) PollTimeout() time.Duration {
	return 60 * time.Second
}
