package storage

import (
	"context"
	"grpc-usdt-rate/internal/models"
)

type Storage interface {
	SaveRate(ctx context.Context, rate models.Rate) error
	Close()
}