package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	BotToken        string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBSSLMode       string
	WebhookURL      string
	WebhookPort     string
	LongPolling     bool
	LogLevel        string
	JWTSecret       string
	JWTIssuer       string
	JWTAudience     string
	FrontendOrigins []string

	// Demo mode (optional second PostgreSQL database)
	DemoEnabled    bool
	DemoDBHost     string
	DemoDBPort     string
	DemoDBUser     string
	DemoDBPassword string
	DemoDBName     string
	DemoDBSSLMode  string

	// Telegram Web Login (OIDC)
	TelegramLoginClientID     string
	TelegramLoginClientSecret string
	TelegramLoginIssuer       string
	TelegramLoginJWKSURL      string
	TelegramLoginNonceTTL     time.Duration
	RegistrationTokenTTL      time.Duration
}

// Load reads configuration from environment variables and returns a Config.
func Load() *Config {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	return &Config{
		BotToken:    os.Getenv("BOT_TOKEN"),
		DBHost:      dbHost,
		DBPort:      dbPort,
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  os.Getenv("DB_PASSWORD"),
		DBName:      getEnv("DB_NAME", "poker_club"),
		DBSSLMode:   dbSSLMode,
		WebhookURL:  os.Getenv("WEBHOOK_URL"),
		WebhookPort: getEnv("WEBHOOK_PORT", "8080"),
		LongPolling: getEnvBool("LONG_POLLING", true),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		JWTIssuer:   getEnv("JWT_ISSUER", "poker-club"),
		JWTAudience: getEnv("JWT_AUDIENCE", "poker-club-web"),
		FrontendOrigins: getEnvSlice(
			"FRONTEND_ORIGINS",
			[]string{
				"http://localhost:3000",
				"http://localhost:3001",
				"https://washstand-lankiness-emcee.ngrok-free.dev", // ngrok: forwarding ngrok -> http://localhost:3000
			},
		),
		DemoEnabled:    getEnvBool("DEMO_ENABLED", false),
		DemoDBHost:     getEnv("DEMO_DB_HOST", dbHost),
		DemoDBPort:     getEnv("DEMO_DB_PORT", dbPort),
		DemoDBUser:     getEnv("DEMO_DB_USER", getEnv("DB_USER", "postgres")),
		DemoDBPassword: firstNonEmpty(os.Getenv("DEMO_DB_PASSWORD"), os.Getenv("DB_PASSWORD")),
		DemoDBName:     getEnv("DEMO_DB_NAME", "poker_club_demo"),
		DemoDBSSLMode:  getEnv("DEMO_DB_SSLMODE", dbSSLMode),
		TelegramLoginClientID:     os.Getenv("TELEGRAM_LOGIN_CLIENT_ID"),
		TelegramLoginClientSecret: os.Getenv("TELEGRAM_LOGIN_CLIENT_SECRET"),
		TelegramLoginIssuer:       getEnv("TELEGRAM_LOGIN_ISSUER", "https://oauth.telegram.org"),
		TelegramLoginJWKSURL:      getEnv("TELEGRAM_LOGIN_JWKS_URL", "https://oauth.telegram.org/.well-known/jwks.json"),
		TelegramLoginNonceTTL:     getEnvDuration("TELEGRAM_LOGIN_NONCE_TTL", 5*time.Minute),
		RegistrationTokenTTL:      getEnvDuration("REGISTRATION_TOKEN_TTL", 15*time.Minute),
	}
}

func getEnvSlice(key string, fallback []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parts := strings.Split(value, ",")

	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	if len(result) == 0 {
		return fallback
	}

	return result
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

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
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

// DemoDSN returns the PostgreSQL connection string for the demo database.
func (c *Config) DemoDSN() string {
	return "host=" + c.DemoDBHost +
		" port=" + c.DemoDBPort +
		" user=" + c.DemoDBUser +
		" password=" + c.DemoDBPassword +
		" dbname=" + c.DemoDBName +
		" sslmode=" + c.DemoDBSSLMode +
		" pool_max_conns=5"
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// WebhookPortDuration returns the webhook port with a colon prefix for Listen.
func (c *Config) WebhookAddr() string {
	return ":" + c.WebhookPort
}

// PollTimeout returns the long polling timeout duration.
func (c *Config) PollTimeout() time.Duration {
	return 60 * time.Second
}
