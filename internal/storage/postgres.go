package storage

import (
	"context"
	"time"
	"grpc-usdt-rate/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(dsn string) (*Postgres, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &Postgres{pool: pool}, nil
}

func (s *Postgres) Close() {
	s.pool.Close()
}

func (s *Postgres) SaveRate(ctx context.Context, rate models.Rate) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO rates (ask, bid, timestamp) VALUES ($1, $2, to_timestamp($3))`,
		rate.Ask, rate.Bid, rate.Timestamp)
	return err
}

