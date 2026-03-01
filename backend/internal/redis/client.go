package redis

import (
	"context"
	"time"

	"yardpass/internal/config"
	"yardpass/internal/observability/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Client struct {
	rdb        *redis.Client
	logger     *zap.Logger
	opDuration *prometheus.HistogramVec
	opErrors   *prometheus.CounterVec
	cacheTotal *prometheus.CounterVec
}

func NewClient(lf fx.Lifecycle, cfg config.RedisConfig, logger *zap.Logger, m *metrics.Metrics) (*Client, error) {
	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(opt)

	opDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "yardpass_redis",
			Name:      "operation_duration_seconds",
			Help:      "Duration of Redis operations",
		},
		[]string{"operation"},
	)

	opErrors := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "yardpass_redis",
			Name:      "operation_errors_total",
			Help:      "Total number of Redis operation errors",
		},
		[]string{"operation"},
	)

	cacheTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "yardpass_redis",
			Name:      "cache_total",
			Help:      "Total number of cache lookups by result (hit or miss)",
		},
		[]string{"result"},
	)

	m.GetRegistry().MustRegister(opDuration, opErrors, cacheTotal)

	lf.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return rdb.Ping(ctx).Err()
		},
		OnStop: func(ctx context.Context) error {
			return rdb.Close()
		},
	})

	return &Client{
		rdb:        rdb,
		logger:     logger,
		opDuration: opDuration,
		opErrors:   opErrors,
		cacheTotal: cacheTotal,
	}, nil
}

func (c *Client) observe(operation string, start time.Time, err error) {
	c.opDuration.WithLabelValues(operation).Observe(time.Since(start).Seconds())
	if err != nil && err != redis.Nil {
		c.opErrors.WithLabelValues(operation).Inc()
	}
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

func (c *Client) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	start := time.Now()
	count, err := c.rdb.Incr(ctx, key).Result()
	c.observe("incr", start, err)
	if err != nil {
		return false, err
	}

	if count == 1 {
		start = time.Now()
		err := c.rdb.Expire(ctx, key, window).Err()
		c.observe("expire", start, err)
		if err != nil {
			return false, err
		}
	}

	return count <= int64(limit), nil
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	start := time.Now()
	result, err := c.rdb.Get(ctx, key).Result()
	c.observe("get", start, err)
	if err == redis.Nil {
		c.cacheTotal.WithLabelValues("miss").Inc()
	} else if err == nil {
		c.cacheTotal.WithLabelValues("hit").Inc()
	}
	return result, err
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	start := time.Now()
	err := c.rdb.Set(ctx, key, value, expiration).Err()
	c.observe("set", start, err)
	return err
}

func (c *Client) Delete(ctx context.Context, key string) error {
	start := time.Now()
	err := c.rdb.Del(ctx, key).Err()
	c.observe("del", start, err)
	return err
}

func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	start := time.Now()
	count, err := c.rdb.Exists(ctx, key).Result()
	c.observe("exists", start, err)
	if err == nil {
		if count > 0 {
			c.cacheTotal.WithLabelValues("hit").Inc()
		} else {
			c.cacheTotal.WithLabelValues("miss").Inc()
		}
	}
	return count > 0, err
}
