package database

import (
	"context"
	"fmt"

	"github.com/dk5761/go-serv/configs"

	"github.com/jackc/pgx/v4/pgxpool"
)

func InitPostgresDB(cfg configs.PostgresConfig) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
	)
	pool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	return pool, nil
}
