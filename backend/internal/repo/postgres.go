package repo

import (
	"context"
	"fmt"

	"yardpass/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type PostgresRepo struct {
	cfg    config.PGConfig
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewPostgresRepo(lf fx.Lifecycle, cfg config.PGConfig, logger *zap.Logger) *PostgresRepo {
	repo := PostgresRepo{
		logger: logger,
		cfg:    cfg,
	}

	lf.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return repo.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return repo.Stop(ctx)
		},
	})

	return &repo
}

func (r *PostgresRepo) Start(ctx context.Context) error {
	config, err := pgxpool.ParseConfig(r.cfg.DSN)
	if err != nil {
		return fmt.Errorf("parse postgres config: %w", err)
	}

	config.MaxConns = int32(r.cfg.MaxConns)
	config.MinConns = int32(r.cfg.MinConns)
	config.MaxConnLifetime = r.cfg.MaxConnLifetime
	config.MaxConnIdleTime = r.cfg.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres pool: %w", err)
	}

	r.pool = pool

	return nil
}

func (r *PostgresRepo) Stop(ctx context.Context) error {
	r.pool.Close()
	return nil
}
