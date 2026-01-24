package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
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
	Level  string `yaml:"level"  env:"LOG_LEVEL"  default:"info"`
	Format string `yaml:"format" env:"LOG_FORMAT" default:"json"`
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

func processValue(v reflect.Value) error {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.CanSet() {
			continue
		}

		if field.Kind() == reflect.Struct && fieldType.Type != reflect.TypeOf(time.Duration(0)) {
			if err := processValue(field); err != nil {
				return err
			}
			continue
		}

		envKey := fieldType.Tag.Get("env")
		defaultVal := fieldType.Tag.Get("default")

		if isZero(field) && defaultVal != "" {
			if err := setFieldValue(field, defaultVal); err != nil {
				return fmt.Errorf("set default for %s: %w", fieldType.Name, err)
			}
		}

		if envKey != "" {
			if envVal := os.Getenv(envKey); envVal != "" {
				if err := setFieldValue(field, envVal); err != nil {
					return fmt.Errorf("set env %s for %s: %w", envKey, fieldType.Name, err)
				}
			}
		}
	}

	return nil
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	}
}

func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("parse duration %q: %w", value, err)
			}
			field.SetInt(int64(d))
			return nil
		}
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("parse int %q: %w", value, err)
		}
		field.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("parse uint %q: %w", value, err)
		}
		field.SetUint(u)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("parse float %q: %w", value, err)
		}
		field.SetFloat(f)

	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("parse bool %q: %w", value, err)
		}
		field.SetBool(b)

	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			parts := strings.Split(value, ",")
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			field.Set(reflect.ValueOf(parts))
		} else {
			return fmt.Errorf("unsupported slice type: %s", field.Type())
		}

	default:
		return fmt.Errorf("unsupported type: %s", field.Kind())
	}

	return nil
}
