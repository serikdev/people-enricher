package database

import (
	"context"
	"fmt"
	"people-enricher/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

func NewPool(ctx context.Context, cfg *config.DBCfg, logger *logrus.Entry) (*pgxpool.Pool, error) {
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
	}).Debug("Connecting database")

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		logger.WithError(err).Error("Failed to parse config")
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.WithError(err).Error("Failed to pool DB")
		return nil, fmt.Errorf("creating pool connection: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		logger.WithError(err).Error("Failed to checking connect")
		return nil, fmt.Errorf("checking connection db: %w", err)
	}
	logger.Info("successfully connecting db using with pgx")
	return pool, nil
}
