package db

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	appconfig "github.com/satryacode/nexth-dummy-be/config"
)

func buildAuthToken(cfg *appconfig.Config) (string, error) {
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.AWSRegion))
	if err != nil {
		return "", fmt.Errorf("unable to load AWS config: %w", err)
	}
	endpoint := fmt.Sprintf("%s:%s", cfg.DBHost, cfg.DBPort)
	token, err := auth.BuildAuthToken(context.Background(), endpoint, cfg.AWSRegion, cfg.DBUser, awsCfg.Credentials)
	if err != nil {
		return "", fmt.Errorf("unable to build auth token: %w", err)
	}
	return token, nil
}

func NewPool(cfg *appconfig.Config) (*pgxpool.Pool, error) {
	token, err := buildAuthToken(cfg)
	if err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=require",
		cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser, token,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to parse pool config: %w", err)
	}

	// Refresh token on each new connection (tokens expire in 15 min)
	poolConfig.BeforeConnect = func(ctx context.Context, connConfig *pgx.ConnConfig) error {
		t, err := buildAuthToken(cfg)
		if err != nil {
			return err
		}
		connConfig.Password = t
		return nil
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create pgx pool: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping db: %w", err)
	}
	return pool, nil
}
