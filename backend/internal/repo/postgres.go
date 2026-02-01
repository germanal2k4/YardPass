package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"yardpass/internal/config"
	"yardpass/internal/observability/logger"
	"yardpass/internal/observability/tracer"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type PostgresRepo struct {
	cfg    config.PGConfig
	pool   *pgxpool.Pool
	logger *zap.Logger
	t      *tracer.Tracer
}

func NewPostgresRepo(lf fx.Lifecycle, cfg config.PGConfig, logger *zap.Logger, t *tracer.Tracer) *PostgresRepo {
	repo := PostgresRepo{
		logger: logger,
		cfg:    cfg,
		t:      t,
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

	queriesToHide := map[string]struct{}{}
	for _, query := range r.cfg.QueriesToHide {
		queriesToHide[query] = struct{}{}
	}

	config.ConnConfig.Tracer = &connTracerWrapper{
		t:              r.t,
		queriesToHide:  queriesToHide,
		fallbackLogger: r.logger,
	}

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

type connTracerWrapper struct {
	t              *tracer.Tracer
	queriesToHide  map[string]struct{}
	fallbackLogger *zap.Logger
}

type queryNameKeyType string

const (
	queryNameKey queryNameKeyType = "queryName"
	startTimeKey queryNameKeyType = "startTime"
)

func (tw *connTracerWrapper) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	if _, ok := tw.queriesToHide[queryNameFromContext(ctx)]; ok {
		return ctx
	}

	ctx, span := tw.t.StartSpan(ctx, "SQL operation")

	lgr := logger.FromContext(ctx)
	if lgr == nil {
		lgr = tw.fallbackLogger
	}

	queryField := zap.String("query", data.SQL)
	span.SetAttributes(attribute.String("query", data.SQL))

	argsS, err := json.Marshal(data.Args)
	if err != nil {
		lgr.Warn("Failed to marshal query args", queryField, zap.Error(err))
	}

	if len(argsS) == 0 {
		argsS = []byte("[]")
	}

	argsField := zap.String("args", string(argsS))
	span.SetAttributes(attribute.String("args", string(argsS)))
	span.SetAttributes(attribute.String("query_name", queryNameFromContext(ctx)))

	ctx = logger.ToContext(ctx, lgr.With(queryField, argsField, zap.String("query_name", queryNameFromContext(ctx))))

	start := time.Now()
	return context.WithValue(ctx, startTimeKey, start)
}

func (tw *connTracerWrapper) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	duration := time.Since(ctx.Value(startTimeKey).(time.Time))
	span.SetAttributes(attribute.Int64("duration_ms", duration.Milliseconds()))

	lgr := logger.FromContext(ctx)
	if lgr == nil {
		lgr = tw.fallbackLogger
	}

	errField := zap.Skip()
	if data.Err != nil {
		errField = zap.Error(data.Err)
		span.SetStatus(codes.Error, data.Err.Error())
		span.RecordError(data.Err)
	}

	span.SetAttributes(attribute.String("command_tag", data.CommandTag.String()))

	lgr.Info("Finished SQL operation", zap.Duration("duration", duration), errField, zap.String("command_tag", data.CommandTag.String()))
}

func queryNameToContext(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, queryNameKey, name)
}

func queryNameFromContext(ctx context.Context) string {
	if v := ctx.Value(queryNameKey); v != nil {
		return v.(string)
	}

	return ""
}
