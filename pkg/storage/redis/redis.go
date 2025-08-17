package redis

import (
	"DeBlockTest/internal/config"
	"context"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/tel-io/tel/v2"
)

const (
	dialTimeout  = 10 * time.Second
	readTimeout  = 3 * time.Second
	writeTimeout = 3 * time.Second
	minIdleConns = 3
	idleTimeout  = 240 * time.Second
)

type Client struct {
	client redis.UniversalClient
}

func Create(ctx context.Context, cfg *config.RedisConfig) (*Client, error) {
	connect, err := createUniversal(ctx, cfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &Client{client: connect}, nil
}

func createUniversal(ctx context.Context, cfg *config.RedisConfig) (redis.UniversalClient, error) {
	hosts := strings.Split(cfg.Addr, ",")

	client := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:        hosts,
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		MinIdleConns: minIdleConns,
		IdleTimeout:  idleTimeout,
		Password:     cfg.Password,
		DB:           cfg.DB,
	})

	if err := ping(ctx, client); err != nil {
		return nil, errors.WithStack(err)
	}

	tel.Global().Info("Redis client connected successfully",
		tel.String("addr", cfg.Addr),
		tel.Int("db", cfg.DB))

	go func() {
		for {
			time.Sleep(2 * time.Second)
			ping(ctx, client)
		}
	}()

	return client, nil
}

func ping(ctx context.Context, client redis.UniversalClient) error {
	err := client.Ping(ctx).Err()
	if err != nil {
		tel.FromCtx(ctx).Error("Redis ping failed", tel.Error(err))
	}
	return err
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) Client() redis.UniversalClient {
	return c.client
}

func (c *Client) GetString(ctx context.Context, key string) (string, error) {
	value, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrNotFound
		}
		return "", errors.Wrap(err, "failed to get string from cache")
	}
	return value, nil
}

func (c *Client) SetString(ctx context.Context, key string, value string, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

var ErrNotFound = errors.New("not found in cache")
