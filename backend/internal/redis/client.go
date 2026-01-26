package redis

import (
	"context"
	"time"

	"yardpass/internal/config"

	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Client struct {
	rdb    *redis.Client
	logger *zap.Logger
}

func NewClient(lf fx.Lifecycle, cfg config.RedisConfig, logger *zap.Logger) (*Client, error) {
	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(opt)

	lf.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return rdb.Ping(ctx).Err()
		},
		OnStop: func(ctx context.Context) error {
			return rdb.Close()
		},
	})

	return &Client{
		rdb:    rdb,
		logger: logger,
	}, nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

func (c *Client) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	count, err := c.rdb.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		if err := c.rdb.Expire(ctx, key, window).Err(); err != nil {
			return false, err
		}
	}

	return count <= int64(limit), nil
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.rdb.Set(ctx, key, value, expiration).Err()
}

func (c *Client) Delete(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}

func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.rdb.Exists(ctx, key).Result()
	return count > 0, err
}
