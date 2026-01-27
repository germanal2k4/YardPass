package logger

import (
	"context"
	"fmt"
	"yardpass/internal/config"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	fallbackLogger *zap.SugaredLogger
)

func init() {
	fallbackLogger = newFallBackLogger()
}

func FallbackLogger() *zap.SugaredLogger {
	return fallbackLogger
}

func NewLogger(lf fx.Lifecycle, cfg config.LogConfig) (*zap.Logger, error) {
	var (
		cores []zapcore.Core
		stops []func()
		info  *loggerInfo
	)

	info = enrichLoggerInfo(cfg)

	stdoutTransport := getStdoutTransport(info)
	cores = append(cores, stdoutTransport.core)
	stops = append(stops, stdoutTransport.stop)

	if cfg.Transport == fileTransport {
		fileTransport, err := getFileTransport(info)
		if err != nil {
			return nil, fmt.Errorf("failed to get file transport for logger: %w", err)
		}

		cores = append(cores, fileTransport.core)
		stops = append(stops, fileTransport.stop)
	}

	if cfg.Transport == elasticTransport {
		elasticTransport, err := getElasticTransport(info)
		if err != nil {
			return nil, fmt.Errorf("failed to get elastic transport for logger: %w", err)
		}

		cores = append(cores, elasticTransport.core)
		stops = append(stops, elasticTransport.stop)
	}

	if len(cores) == 0 {
		return nil, fmt.Errorf("no logger could be created for %s", cfg.Transport)
	}

	lgr := zap.New(zapcore.NewTee(cores...))

	lf.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			for _, stop := range stops {
				stop()
			}
			return nil
		},
	})

	return lgr, nil
}

func NewNopLogger() *zap.Logger {
	return zap.NewNop()
}

func newFallBackLogger() *zap.SugaredLogger {
	fallbackCfg := config.LogConfig{
		Level:  "info",
		Format: "json",
	}

	info := enrichLoggerInfo(fallbackCfg)

	stdoutTransport := getStdoutTransport(info)
	lgr := zap.New(stdoutTransport.core).Sugar()
	return lgr
}

type loggerInfo struct {
	cfg    config.LogConfig
	encCfg zapcore.EncoderConfig
	lvl    zap.AtomicLevel
}

func enrichLoggerInfo(cfg config.LogConfig) *loggerInfo {
	info := &loggerInfo{
		cfg: cfg,
	}

	info.encCfg = zap.NewProductionEncoderConfig()
	info.encCfg.EncodeTime = zapcore.RFC3339TimeEncoder

	info.lvl = zap.NewAtomicLevel()
	if err := info.lvl.UnmarshalText([]byte(cfg.Level)); err != nil {
		info.lvl = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	return info
}

type loggerKey string

func ToContext(ctx context.Context, lgr *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, loggerKey("logger"), lgr)
}

func FromContext(ctx context.Context) *zap.SugaredLogger {
	if ctx == nil {
		return nil
	}
	lgr, ok := ctx.Value(loggerKey("logger")).(*zap.SugaredLogger)
	if !ok {
		return nil
	}
	return lgr
}
