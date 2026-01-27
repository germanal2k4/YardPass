package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear all env vars that might affect the test
	clearEnvVars(t)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Server defaults
	assertEqual(t, "Server.Host", "0.0.0.0", cfg.Server.Host)
	assertEqual(t, "Server.Port", "8080", cfg.Server.Port)
	assertEqual(t, "Server.StartTimeout", 15*time.Second, cfg.Server.StartTimeout)
	assertEqual(t, "Server.StopTimeout", 15*time.Second, cfg.Server.StopTimeout)

	// PG defaults
	assertEqual(t, "PG.DSN", "postgres://yardpass:password@localhost:5432/yardpass?sslmode=disable", cfg.PG.DSN)
	assertEqual(t, "PG.MaxConns", 25, cfg.PG.MaxConns)
	assertEqual(t, "PG.MinConns", 5, cfg.PG.MinConns)
	assertEqual(t, "PG.MaxConnLifetime", time.Hour, cfg.PG.MaxConnLifetime)
	assertEqual(t, "PG.MaxConnIdleTime", 30*time.Minute, cfg.PG.MaxConnIdleTime)

	// Redis defaults
	assertEqual(t, "Redis.URL", "redis://localhost:6379/0", cfg.Redis.URL)

	// JWT defaults
	assertEqual(t, "JWT.Secret", "", cfg.JWT.Secret)
	assertEqual(t, "JWT.AccessTTL", 15*time.Minute, cfg.JWT.AccessTTL)
	assertEqual(t, "JWT.RefreshTTL", 168*time.Hour, cfg.JWT.RefreshTTL)

	// Telegram defaults
	assertEqual(t, "Telegram.BotToken", "", cfg.Telegram.BotToken)
	assertEqual(t, "Telegram.WebhookURL", "", cfg.Telegram.WebhookURL)
	assertEqual(t, "Telegram.ServerHost", "0.0.0.0", cfg.Telegram.ServerHost)
	assertEqual(t, "Telegram.ServerPort", "8081", cfg.Telegram.ServerPort)

	// RateLimit defaults
	assertEqual(t, "RateLimit.RequestsPerMinute", 60, cfg.RateLimit.RequestsPerMinute)
	assertEqual(t, "RateLimit.CreatePassPerHour", 10, cfg.RateLimit.CreatePassPerHour)
	assertEqual(t, "RateLimit.ScanPerMinute", 100, cfg.RateLimit.ScanPerMinute)
}

func TestLoad_FromYAML(t *testing.T) {
	clearEnvVars(t)

	yamlContent := `
server:
  host: "192.168.1.1"
  port: "9000"
  start_timeout: 30s
  stop_timeout: 20s

pg:
  dsn: "postgres://user:pass@db:5432/mydb"
  max_conns: 50
  min_conns: 10

redis:
  url: "redis://redis:6379/1"

jwt:
  secret: "yaml-secret"
  access_ttl: 30m
  refresh_ttl: 720h

telegram:
  bot_token: "yaml-bot-token"
  webhook_url: "https://example.com/webhook"

rate_limit:
  requests_per_minute: 120
  create_pass_per_hour: 20

log:
  level: "debug"
  format: "text"
`
	configPath := createTempConfig(t, yamlContent)

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Server from YAML
	assertEqual(t, "Server.Host", "192.168.1.1", cfg.Server.Host)
	assertEqual(t, "Server.Port", "9000", cfg.Server.Port)
	assertEqual(t, "Server.StartTimeout", 30*time.Second, cfg.Server.StartTimeout)
	assertEqual(t, "Server.StopTimeout", 20*time.Second, cfg.Server.StopTimeout)

	// PG from YAML
	assertEqual(t, "PG.DSN", "postgres://user:pass@db:5432/mydb", cfg.PG.DSN)
	assertEqual(t, "PG.MaxConns", 50, cfg.PG.MaxConns)
	assertEqual(t, "PG.MinConns", 10, cfg.PG.MinConns)
	// These should be defaults since not in YAML
	assertEqual(t, "PG.MaxConnLifetime", time.Hour, cfg.PG.MaxConnLifetime)
	assertEqual(t, "PG.MaxConnIdleTime", 30*time.Minute, cfg.PG.MaxConnIdleTime)

	// Redis from YAML
	assertEqual(t, "Redis.URL", "redis://redis:6379/1", cfg.Redis.URL)

	// JWT from YAML
	assertEqual(t, "JWT.Secret", "yaml-secret", cfg.JWT.Secret)
	assertEqual(t, "JWT.AccessTTL", 30*time.Minute, cfg.JWT.AccessTTL)
	assertEqual(t, "JWT.RefreshTTL", 720*time.Hour, cfg.JWT.RefreshTTL)

	// Telegram from YAML
	assertEqual(t, "Telegram.BotToken", "yaml-bot-token", cfg.Telegram.BotToken)
	assertEqual(t, "Telegram.WebhookURL", "https://example.com/webhook", cfg.Telegram.WebhookURL)

	// RateLimit from YAML
	assertEqual(t, "RateLimit.RequestsPerMinute", 120, cfg.RateLimit.RequestsPerMinute)
	assertEqual(t, "RateLimit.CreatePassPerHour", 20, cfg.RateLimit.CreatePassPerHour)
	// Default since not in YAML
	assertEqual(t, "RateLimit.ScanPerMinute", 100, cfg.RateLimit.ScanPerMinute)
}

func TestLoad_EnvOverridesYAML(t *testing.T) {
	clearEnvVars(t)

	yamlContent := `
server:
  host: "yaml-host"
  port: "yaml-port"

pg:
  dsn: "yaml-dsn"
  max_conns: 10

jwt:
  secret: "yaml-secret"
  access_ttl: 10m

log:
  level: "warn"
`
	configPath := createTempConfig(t, yamlContent)

	// Set env vars to override YAML
	t.Setenv("SERVER_HOST", "env-host")
	t.Setenv("SERVER_PORT", "env-port")
	t.Setenv("DATABASE_URL", "env-dsn")
	t.Setenv("PG_MAX_CONNS", "99")
	t.Setenv("JWT_SECRET", "env-secret")
	t.Setenv("JWT_ACCESS_TTL", "1h")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Env should override YAML
	assertEqual(t, "Server.Host", "env-host", cfg.Server.Host)
	assertEqual(t, "Server.Port", "env-port", cfg.Server.Port)
	assertEqual(t, "PG.DSN", "env-dsn", cfg.PG.DSN)
	assertEqual(t, "PG.MaxConns", 99, cfg.PG.MaxConns)
	assertEqual(t, "JWT.Secret", "env-secret", cfg.JWT.Secret)
	assertEqual(t, "JWT.AccessTTL", time.Hour, cfg.JWT.AccessTTL)
}

func TestLoad_EnvOverridesDefaults(t *testing.T) {
	clearEnvVars(t)

	t.Setenv("SERVER_HOST", "custom-host")
	t.Setenv("SERVER_PORT", "3000")
	t.Setenv("SERVER_START_TIMEOUT", "45s")
	t.Setenv("DATABASE_URL", "postgres://custom:pass@localhost/db")
	t.Setenv("REDIS_URL", "redis://custom:6379")
	t.Setenv("JWT_SECRET", "super-secret")
	t.Setenv("JWT_ACCESS_TTL", "2h")
	t.Setenv("JWT_REFRESH_TTL", "336h")
	t.Setenv("TELEGRAM_BOT_TOKEN", "bot-token-123")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	assertEqual(t, "Server.Host", "custom-host", cfg.Server.Host)
	assertEqual(t, "Server.Port", "3000", cfg.Server.Port)
	assertEqual(t, "Server.StartTimeout", 45*time.Second, cfg.Server.StartTimeout)
	assertEqual(t, "PG.DSN", "postgres://custom:pass@localhost/db", cfg.PG.DSN)
	assertEqual(t, "Redis.URL", "redis://custom:6379", cfg.Redis.URL)
	assertEqual(t, "JWT.Secret", "super-secret", cfg.JWT.Secret)
	assertEqual(t, "JWT.AccessTTL", 2*time.Hour, cfg.JWT.AccessTTL)
	assertEqual(t, "JWT.RefreshTTL", 336*time.Hour, cfg.JWT.RefreshTTL)
	assertEqual(t, "Telegram.BotToken", "bot-token-123", cfg.Telegram.BotToken)
}

func TestLoad_InvalidYAML(t *testing.T) {
	clearEnvVars(t)

	yamlContent := `
server:
  host: [invalid yaml
`
	configPath := createTempConfig(t, yamlContent)

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("Load() expected error for invalid YAML, got nil")
	}
}

func TestLoad_NonExistentConfigFile(t *testing.T) {
	clearEnvVars(t)

	// Non-existent file should not error, just use defaults
	cfg, err := Load("/non/existent/path/config.yaml")
	if err != nil {
		t.Fatalf("Load() error = %v, want nil for non-existent file", err)
	}

	// Should have defaults
	assertEqual(t, "Server.Host", "0.0.0.0", cfg.Server.Host)
	assertEqual(t, "Server.Port", "8080", cfg.Server.Port)
}

func TestLoad_EmptyConfigFile(t *testing.T) {
	clearEnvVars(t)

	configPath := createTempConfig(t, "")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Should have defaults
	assertEqual(t, "Server.Host", "0.0.0.0", cfg.Server.Host)
	assertEqual(t, "Server.Port", "8080", cfg.Server.Port)
}

func TestLoad_PartialYAML(t *testing.T) {
	clearEnvVars(t)

	yamlContent := `
server:
  host: "partial-host"
# port not specified, should use default

pg:
  max_conns: 100
# dsn not specified, should use default
`
	configPath := createTempConfig(t, yamlContent)

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// From YAML
	assertEqual(t, "Server.Host", "partial-host", cfg.Server.Host)
	assertEqual(t, "PG.MaxConns", 100, cfg.PG.MaxConns)

	// Defaults for missing fields
	assertEqual(t, "Server.Port", "8080", cfg.Server.Port)
	assertEqual(t, "PG.DSN", "postgres://yardpass:password@localhost:5432/yardpass?sslmode=disable", cfg.PG.DSN)
}

func TestLoad_InvalidDuration(t *testing.T) {
	clearEnvVars(t)

	t.Setenv("SERVER_START_TIMEOUT", "invalid-duration")

	_, err := Load("")
	if err == nil {
		t.Fatal("Load() expected error for invalid duration, got nil")
	}
}

func TestLoad_InvalidInt(t *testing.T) {
	clearEnvVars(t)

	t.Setenv("PG_MAX_CONNS", "not-a-number")

	_, err := Load("")
	if err == nil {
		t.Fatal("Load() expected error for invalid int, got nil")
	}
}

func TestLoad_PriorityOrder(t *testing.T) {
	// Test: env > yaml > default
	clearEnvVars(t)

	yamlContent := `
server:
  host: "yaml-value"
  port: "yaml-port"
`
	configPath := createTempConfig(t, yamlContent)

	// Only override host with env
	t.Setenv("SERVER_HOST", "env-value")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// env overrides yaml
	assertEqual(t, "Server.Host", "env-value", cfg.Server.Host)
	// yaml value preserved
	assertEqual(t, "Server.Port", "yaml-port", cfg.Server.Port)
	// default for unspecified
	assertEqual(t, "Server.StartTimeout", 15*time.Second, cfg.Server.StartTimeout)
}

func TestSetFieldValue_AllTypes(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		env     map[string]string
		check   func(*Config) error
		wantErr bool
	}{
		{
			name: "duration_seconds",
			env:  map[string]string{"SERVER_START_TIMEOUT": "45s"},
			check: func(cfg *Config) error {
				if cfg.Server.StartTimeout != 45*time.Second {
					return errorf("StartTimeout = %v, want 45s", cfg.Server.StartTimeout)
				}
				return nil
			},
		},
		{
			name: "duration_minutes",
			env:  map[string]string{"JWT_ACCESS_TTL": "30m"},
			check: func(cfg *Config) error {
				if cfg.JWT.AccessTTL != 30*time.Minute {
					return errorf("AccessTTL = %v, want 30m", cfg.JWT.AccessTTL)
				}
				return nil
			},
		},
		{
			name: "duration_hours",
			env:  map[string]string{"PG_MAX_CONN_LIFETIME": "2h"},
			check: func(cfg *Config) error {
				if cfg.PG.MaxConnLifetime != 2*time.Hour {
					return errorf("MaxConnLifetime = %v, want 2h", cfg.PG.MaxConnLifetime)
				}
				return nil
			},
		},
		{
			name: "int_value",
			env:  map[string]string{"PG_MAX_CONNS": "42"},
			check: func(cfg *Config) error {
				if cfg.PG.MaxConns != 42 {
					return errorf("MaxConns = %d, want 42", cfg.PG.MaxConns)
				}
				return nil
			},
		},
		{
			name: "string_value",
			env:  map[string]string{"JWT_SECRET": "my-secret-key"},
			check: func(cfg *Config) error {
				if cfg.JWT.Secret != "my-secret-key" {
					return errorf("Secret = %s, want my-secret-key", cfg.JWT.Secret)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars(t)

			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			var configPath string
			if tt.yaml != "" {
				configPath = createTempConfig(t, tt.yaml)
			}

			cfg, err := Load(configPath)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && tt.check != nil {
				if checkErr := tt.check(cfg); checkErr != nil {
					t.Error(checkErr)
				}
			}
		})
	}
}

// Helper functions

func clearEnvVars(t *testing.T) {
	t.Helper()
	envVars := []string{
		"SERVER_HOST", "SERVER_PORT", "SERVER_START_TIMEOUT", "SERVER_STOP_TIMEOUT",
		"DATABASE_URL", "PG_MAX_CONNS", "PG_MIN_CONNS", "PG_MAX_CONN_LIFETIME", "PG_MAX_CONN_IDLE_TIME",
		"REDIS_URL",
		"JWT_SECRET", "JWT_ACCESS_TTL", "JWT_REFRESH_TTL",
		"TELEGRAM_BOT_TOKEN", "TELEGRAM_WEBHOOK_URL", "TELEGRAM_SERVER_HOST", "TELEGRAM_SERVER_PORT",
		"SERVICE_TOKEN",
		"RATE_LIMIT_REQUESTS_PER_MINUTE", "RATE_LIMIT_CREATE_PASS_PER_HOUR", "RATE_LIMIT_SCAN_PER_MINUTE",
		"LOG_LEVEL", "LOG_FORMAT",
	}
	for _, v := range envVars {
		t.Setenv(v, "")
	}
}

func createTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp config: %v", err)
	}
	return path
}

func assertEqual[T comparable](t *testing.T, name string, want, got T) {
	t.Helper()
	if want != got {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

func errorf(format string, args ...any) error {
	return &testError{msg: format, args: args}
}

type testError struct {
	msg  string
	args []any
}

func (e *testError) Error() string {
	return e.msg
}
