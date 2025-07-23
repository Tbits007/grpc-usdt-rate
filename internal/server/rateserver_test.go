package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"grpc-usdt-rate/internal/models"

	pb "grpc-usdt-rate/api/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockStorage struct {
	SaveFunc  func(ctx context.Context, rate models.Rate) error
	CloseFunc func()
}

func (m *mockStorage) SaveRate(ctx context.Context, rate models.Rate) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, rate)
	}
	return nil
}

func (m *mockStorage) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}

func TestHealthCheck(t *testing.T) {
	srv := NewRateServer(&mockStorage{}, zap.NewNop())
	resp, err := srv.HealthCheck(context.Background(), &pb.Empty{})
	require.NoError(t, err)
	assert.True(t, resp.Status)
}

func TestGetRates_Success(t *testing.T) {
	grinex := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
			"asks": [{"price": "100.5", "volume": "10"}],
			"bids": [{"price": "99.5", "volume": "20"}]
		}`))
	}))
	defer grinex.Close()

	mockDB := &mockStorage{
		SaveFunc: func(ctx context.Context, rate models.Rate) error {
			return nil
		},
	}

	srv := NewRateServer(mockDB, zap.NewNop())
	srv.Client = &http.Client{Timeout: 5 * time.Second}

	resp, err := srv.GetRates(context.Background(), &pb.Empty{})
	require.NoError(t, err)
	assert.NotZero(t, resp.Timestamp)
}

func TestGetRates_SaveError(t *testing.T) {
	grinex := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
			"asks": [{"price": "100.5", "volume": "10"}],
			"bids": [{"price": "99.5", "volume": "20"}]
		}`))
	}))
	defer grinex.Close()

	mockDB := &mockStorage{
		SaveFunc: func(ctx context.Context, rate models.Rate) error {
			return errors.New("failed to save")
		},
	}

	srv := NewRateServer(mockDB, zap.NewNop())
	srv.Client = &http.Client{Timeout: 5 * time.Second}

	_, err := srv.GetRates(context.Background(), &pb.Empty{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save rate failed")
}