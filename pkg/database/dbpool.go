package repository

import (
	"context"
	"fmt"
	"people-enricher/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

func NewPool(ctx context.Context, cfg *config.DBCfg, logger *logrus.Logger) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
	)

	logger.WithFields(logrus.Fields{
		"host":     cfg.Host,
		"port":     cfg.Port,
		"dbname":   cfg.DBName,
		"sslmode":  cfg.SSLMode,
		"username": cfg.User,
	}).Debug("Connecting data base")

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxConnection)
	poolConfig.MaxConns = int32(cfg.IdleConnection)

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("creating pool connection")
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("checking connection db")
	}
	logger.Info("successfully connecting db using with pgx")
	return pool, nil
}
