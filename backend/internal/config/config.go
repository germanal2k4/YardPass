package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	Telegram  TelegramConfig
	Service   ServiceConfig
	RateLimit RateLimitConfig
	Log       LogConfig
}

type ServerConfig struct {
	Host string
	Port string
}

type DatabaseConfig struct {
	URL string
}

type RedisConfig struct {
	URL string
}

type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type TelegramConfig struct {
	BotToken   string
	WebhookURL string
}

type ServiceConfig struct {
	Token string
}

type RateLimitConfig struct {
	RequestsPerMinute int
	CreatePassPerHour int
	ScanPerMinute     int
}

type LogConfig struct {
	Level  string
	Format string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "postgres://yardpass:password@localhost:5432/yardpass?sslmode=disable"),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "redis://localhost:6379/0"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", ""),
			AccessTTL:  parseDuration(getEnv("JWT_ACCESS_TTL", "15m")),
			RefreshTTL: parseDuration(getEnv("JWT_REFRESH_TTL", "168h")),
		},
		Telegram: TelegramConfig{
			BotToken:   getEnv("TELEGRAM_BOT_TOKEN", ""),
			WebhookURL: getEnv("TELEGRAM_WEBHOOK_URL", ""),
		},
		Service: ServiceConfig{
			Token: getEnv("SERVICE_TOKEN", ""),
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute: parseInt(getEnv("RATE_LIMIT_REQUESTS_PER_MINUTE", "60")),
			CreatePassPerHour: parseInt(getEnv("RATE_LIMIT_CREATE_PASS_PER_HOUR", "10")),
			ScanPerMinute:     parseInt(getEnv("RATE_LIMIT_SCAN_PER_MINUTE", "100")),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}


	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 15 * time.Minute
	}
	return d
}
