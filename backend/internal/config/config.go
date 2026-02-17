package config

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/joho/godotenv"
)

// Config is the main application configuration.
// Tags:
//   - yaml:"name"       — field name in YAML config
//   - env:"VAR_NAME"    — environment variable name (overrides yaml)
//   - default:"value"   — default value if not set in yaml or env
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	PG        PGConfig        `yaml:"pg"`
	Redis     RedisConfig     `yaml:"redis"`
	JWT       JWTConfig       `yaml:"jwt"`
	Telegram  TelegramConfig  `yaml:"telegram"`
	Service   ServiceConfig   `yaml:"service"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	Log       LogConfig       `yaml:"log"`
	Tracer    TracerConfig    `yaml:"tracer"`
	Metrics   MetricsConfig   `yaml:"metrics"`
}

type ServerConfig struct {
	Host         string        `yaml:"host"          env:"SERVER_HOST"          default:"0.0.0.0"`
	Port         string        `yaml:"port"          env:"SERVER_PORT"          default:"8080"`
	StartTimeout time.Duration `yaml:"start_timeout" env:"SERVER_START_TIMEOUT" default:"15s"`
	StopTimeout  time.Duration `yaml:"stop_timeout"  env:"SERVER_STOP_TIMEOUT"  default:"15s"`
}

type PGConfig struct {
	DSN             string        `yaml:"dsn"                env:"DATABASE_URL"            default:"postgres://yardpass:password@localhost:5432/yardpass?sslmode=disable"`
	MaxConns        int           `yaml:"max_conns"          env:"PG_MAX_CONNS"            default:"25"`
	MinConns        int           `yaml:"min_conns"          env:"PG_MIN_CONNS"            default:"5"`
	MaxConnLifetime time.Duration `yaml:"max_conn_lifetime"  env:"PG_MAX_CONN_LIFETIME"    default:"1h"`
	MaxConnIdleTime time.Duration `yaml:"max_conn_idle_time" env:"PG_MAX_CONN_IDLE_TIME"  default:"30m"`
	QueriesToHide   []string      `yaml:"queries_to_hide"    env:"PG_QUERIES_TO_HIDE"     default:""`
}

type RedisConfig struct {
	URL string `yaml:"url" env:"REDIS_URL" default:"redis://localhost:6379/0"`
}

type JWTConfig struct {
	Secret     string        `yaml:"secret"      env:"JWT_SECRET"      default:""`
	AccessTTL  time.Duration `yaml:"access_ttl"  env:"JWT_ACCESS_TTL"  default:"15m"`
	RefreshTTL time.Duration `yaml:"refresh_ttl" env:"JWT_REFRESH_TTL" default:"168h"`
}

type TelegramConfig struct {
	BotToken   string `yaml:"bot_token"   env:"TELEGRAM_BOT_TOKEN"   default:""`
	WebhookURL string `yaml:"webhook_url" env:"TELEGRAM_WEBHOOK_URL" default:""`
	ServerHost string `yaml:"server_host" env:"TELEGRAM_SERVER_HOST" default:"0.0.0.0"`
	ServerPort string `yaml:"server_port" env:"TELEGRAM_SERVER_PORT" default:"8081"`
}

type ServiceConfig struct {
	Token string `yaml:"token" env:"SERVICE_TOKEN" default:""`
}

type RateLimitConfig struct {
	RequestsPerMinute int `yaml:"requests_per_minute"  env:"RATE_LIMIT_REQUESTS_PER_MINUTE" default:"60"`
	CreatePassPerHour int `yaml:"create_pass_per_hour" env:"RATE_LIMIT_CREATE_PASS_PER_HOUR" default:"10"`
	ScanPerMinute     int `yaml:"scan_per_minute"      env:"RATE_LIMIT_SCAN_PER_MINUTE"     default:"100"`
}

type LogConfig struct {
	Disabled       bool           `yaml:"disabled"         default:"false"`
	Level          string         `yaml:"level"            default:"info"`
	Format         string         `yaml:"format"           default:"json"`
	Buffered       bool           `yaml:"buffered"         default:"false"`
	Transport      string         `yaml:"transport"        default:""`
	FilePath       string         `yaml:"file_path"        default:""`
	ElasticConfig  *ElasticConfig `yaml:"elastic_config"   default:""`
	MaskHeaders    []string       `yaml:"mask_headers"     env:"LOG_MASK_HEADERS"     default:"Authorization,X-Service-Token,Cookie"`
	MaskBodyFields []string       `yaml:"mask_body_fields" env:"LOG_MASK_BODY_FIELDS" default:"password,token,secret,api_key,apiKey,refresh_token,access_token"`
}

type ElasticConfig struct {
	Url             string        `yaml:"url" env:"ELASTIC_URL" default:"http://localhost:9200"`
	Index           string        `yaml:"index" default:"yardpass-logs"`
	FlushInterval   time.Duration `yaml:"flush_interval" default:"1s"`
	WriteBufferSize int           `yaml:"write_buffer_size" default:"1024"`
}

type TracerConfig struct {
	Enabled bool   `yaml:"enabled" default:"true"`
	Url     string `yaml:"url" env:"TRACER_URL" default:"localhost:4317"`
}

type MetricsConfig struct {
	Enabled bool `yaml:"enabled" default:"true"`
	Port    int  `yaml:"port" default:"3000"`
}

func Load(configPath string) (*Config, error) {
	_ = godotenv.Load()

	var cfg Config

	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		if err == nil && len(data) > 0 {
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return nil, fmt.Errorf("unmarshal config: %w", err)
			}
		}
	}

	if err := processConfig(&cfg); err != nil {
		return nil, fmt.Errorf("process config: %w", err)
	}

	return &cfg, nil
}

func processConfig(cfg any) error {
	return processValue(reflect.ValueOf(cfg))
}

func ApplyDefaults(configPath string) error {
	existing := make(map[string]any)
	if data, err := os.ReadFile(configPath); err == nil && len(data) > 0 {
		if err := yaml.Unmarshal(data, &existing); err != nil {
			return fmt.Errorf("unmarshal existing config: %w", err)
		}
	}

	defaults := buildDefaultsMap(reflect.TypeOf(Config{}), reflect.ValueOf(Config{}))

	merged := mergeMaps(defaults, existing)

	output, err := yaml.MarshalWithOptions(merged, yaml.IndentSequence(true), yaml.UseLiteralStyleIfMultiline(true))
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	header := "# YardPass Configuration\n# Auto-generated with default values. Override as needed.\n# Environment variables take precedence over these values.\n\n"

	if err := os.WriteFile(configPath, []byte(header+string(output)), 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}
