package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"yardpass/internal/config"
	"yardpass/internal/observability/logger"
	"yardpass/internal/observability/metrics"
	"yardpass/internal/observability/tracer"
	"yardpass/internal/repo/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/google/uuid"
)

type PostgresRepo struct {
	cfg     config.PGConfig
	pool    *pgxpool.Pool
	queries *db.Queries
	logger  *zap.Logger
	t       *tracer.Tracer
	metrics *metrics.Metrics
}

// NewPostgresRepoFromPool creates a PostgresRepo from an existing pool, for integration tests.
// Caller is responsible for closing the pool. Logger may be nil (uses Nop).
func NewPostgresRepoFromPool(pool *pgxpool.Pool, logger *zap.Logger) *PostgresRepo {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &PostgresRepo{
		pool:    pool,
		queries: db.New(pool),
		logger:  logger,
	}
}

func NewPostgresRepo(lf fx.Lifecycle, cfg config.PGConfig, logger *zap.Logger, t *tracer.Tracer, m *metrics.Metrics) *PostgresRepo {
	repo := PostgresRepo{
		logger:  logger,
		cfg:     cfg,
		t:       t,
		metrics: m,
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

	queryDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "yardpass_db",
			Name:      "query_duration_seconds",
			Help:      "Duration of database queries",
		},
		[]string{"query_name"},
	)

	queryErrors := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "yardpass_db",
			Name:      "query_errors_total",
			Help:      "Total number of database query errors",
		},
		[]string{"query_name"},
	)

	registry := r.metrics.GetRegistry()
	registry.MustRegister(queryDuration, queryErrors)

	config.ConnConfig.Tracer = &connTracerWrapper{
		t:              r.t,
		queriesToHide:  queriesToHide,
		fallbackLogger: r.logger,
		queryDuration:  queryDuration,
		queryErrors:    queryErrors,
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres pool: %w", err)
	}

	r.pool = pool
	r.queries = db.New(pool)

	registry.MustRegister(
		prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace: "yardpass_db",
			Name:      "pool_total_conns",
			Help:      "Total number of connections in the pool",
		}, func() float64 { return float64(r.pool.Stat().TotalConns()) }),
		prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace: "yardpass_db",
			Name:      "pool_idle_conns",
			Help:      "Number of idle connections in the pool",
		}, func() float64 { return float64(r.pool.Stat().IdleConns()) }),
	)

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
	queryDuration  *prometheus.HistogramVec
	queryErrors    *prometheus.CounterVec
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
	startTime, ok := ctx.Value(startTimeKey).(time.Time)
	if !ok {
		return
	}

	span := trace.SpanFromContext(ctx)
	defer span.End()

	duration := time.Since(startTime)
	span.SetAttributes(attribute.Int64("duration_ms", duration.Milliseconds()))

	queryName := queryNameFromContext(ctx)
	tw.queryDuration.WithLabelValues(queryName).Observe(duration.Seconds())
	if data.Err != nil {
		tw.queryErrors.WithLabelValues(queryName).Inc()
	}

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

func uuidToPgtype(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func timeToPgtypeTimestamp(t *time.Time) pgtype.Timestamp {
	if t == nil {
		return pgtype.Timestamp{}
	}
	return pgtype.Timestamp{Time: *t, Valid: true}
}

func intToInt32Ptr(v int) *int32 {
	if v == 0 {
		return nil
	}
	i := int32(v)
	return &i
}
