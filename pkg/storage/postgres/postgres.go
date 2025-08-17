package postgres

import (
	"DeBlockTest/internal/config"
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/tel-io/tel/v2"
)

type Client struct {
	pool *pgxpool.Pool
	cfg  *config.DatabaseConfig
}

func Create(ctx context.Context, cfg *config.DatabaseConfig) (*Client, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse database config")
	}

	poolConfig.MaxConns = int32(cfg.MaxConns)

	pool, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to PostgreSQL")
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, errors.Wrap(err, "failed to ping PostgreSQL")
	}

	client := &Client{
		pool: pool,
		cfg:  cfg,
	}

	tel.Global().Info("PostgreSQL client connected successfully",
		tel.String("host", cfg.Host),
		tel.Int("port", cfg.Port),
		tel.String("database", cfg.DBName),
		tel.Int("max_conns", cfg.MaxConns))

	return client, nil
}

func (p *Client) Close() {
	if p.pool != nil {
		p.pool.Close()
	}
}

func (p *Client) Pool() *pgxpool.Pool {
	return p.pool
}

func (p *Client) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return p.pool.Query(ctx, sql, args...)
}

func (p *Client) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return p.pool.QueryRow(ctx, sql, args...)
}

func (p *Client) Exec(ctx context.Context, sql string, args ...interface{}) error {
	_, err := p.pool.Exec(ctx, sql, args...)
	return err
}
