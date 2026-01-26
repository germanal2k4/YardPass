package observability

import (
	"context"
	"errors"
	"syscall"
	"yardpass/internal/config"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(lf fx.Lifecycle, cfg config.LogConfig) (*zap.Logger, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(cfg.Level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	var config zap.Config
	if cfg.Format == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	zap.ReplaceGlobals(logger)

	lf.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			err := logger.Sync()
			if err != nil && !errors.Is(err, syscall.ENOTTY) {
				return err
			}
			return nil
		},
	})

	return logger, nil
}

func NewNopLogger() *zap.Logger {
	return zap.NewNop()
}
