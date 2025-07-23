package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"grpc-usdt-rate/internal/metrics"
	"grpc-usdt-rate/internal/models"
	"grpc-usdt-rate/internal/storage"

	pb "grpc-usdt-rate/api/proto"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type RateServer struct {
	pb.UnimplementedRateServiceServer
	db     storage.Storage
	log    *zap.Logger
	tracer trace.Tracer
	Client *http.Client
}

func NewRateServer(db storage.Storage, log *zap.Logger) *RateServer {
	return &RateServer{
		db:     db,
		log:    log,
		tracer: otel.Tracer("rate-service"),
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *RateServer) Register(grpcServer *grpc.Server) {
	pb.RegisterRateServiceServer(grpcServer, s)
}

func (s *RateServer) GetRates(ctx context.Context, _ *pb.Empty) (*pb.RateResponse, error) {
    ctx, span := s.tracer.Start(ctx, "GetRates")
    defer span.End()

    timer := prometheus.NewTimer(metrics.RequestDuration.WithLabelValues("GetRates"))
    metrics.RequestsTotal.WithLabelValues("GetRates").Inc()
    defer timer.ObserveDuration()

    s.log.Info("Fetching rate from Grinex")

    depth, err := s.fetchRates(ctx)
    if err != nil {
        metrics.RequestErrors.WithLabelValues("GetRates").Inc()
        s.log.Error("Failed to fetch rates", zap.Error(err))
        span.RecordError(err)
        return nil, fmt.Errorf("fetch rates failed: %w", err)
    }

    if len(depth.Asks) == 0 {
        err := fmt.Errorf("empty asks array")
        metrics.RequestErrors.WithLabelValues("GetRates").Inc()
        s.log.Error("Empty asks array", zap.Error(err))
        span.RecordError(err)
        return nil, err
    }

    if len(depth.Bids) == 0 {
        err := fmt.Errorf("empty bids array")
        metrics.RequestErrors.WithLabelValues("GetRates").Inc()
        s.log.Error("Empty bids array", zap.Error(err))
        span.RecordError(err)
        return nil, err
    }

    ask, err := strconv.ParseFloat(depth.Asks[0].Price, 64)
    if err != nil {
        metrics.RequestErrors.WithLabelValues("GetRates").Inc()
        s.log.Error("Failed to parse ask price", zap.Error(err))
        span.RecordError(err)
        return nil, fmt.Errorf("parse ask price failed: %w", err)
    }

    bid, err := strconv.ParseFloat(depth.Bids[0].Price, 64)
    if err != nil {
        metrics.RequestErrors.WithLabelValues("GetRates").Inc()
        s.log.Error("Failed to parse bid price", zap.Error(err))
        span.RecordError(err)
        return nil, fmt.Errorf("parse bid price failed: %w", err)
    }

    rate := models.Rate{
        Ask:       ask,
        Bid:       bid,
        Timestamp: time.Now().Unix(),
    }

    if err := s.db.SaveRate(ctx, rate); err != nil {
        metrics.RequestErrors.WithLabelValues("SaveRate").Inc()
        s.log.Error("Failed to save rate", zap.Error(err))
        span.RecordError(err)
        return nil, fmt.Errorf("save rate failed: %w", err)
    }

    s.log.Info("Successfully fetched and stored rate",
        zap.Float64("ask", rate.Ask),
        zap.Float64("bid", rate.Bid),
        zap.Int64("timestamp", rate.Timestamp),
    )

    return &pb.RateResponse{
        Ask:       rate.Ask,
        Bid:       rate.Bid,
        Timestamp: rate.Timestamp,
    }, nil
}

func (s *RateServer) HealthCheck(ctx context.Context, _ *pb.Empty) (*pb.HealthResponse, error) {
	metrics.RequestsTotal.WithLabelValues("HealthCheck").Inc()
	return &pb.HealthResponse{Status: true}, nil
}

type DepthResponse struct {
    Asks []struct {
        Price  string `json:"price"`
        Volume string `json:"volume"`
    } `json:"asks"`
    Bids []struct {
        Price  string `json:"price"`
        Volume string `json:"volume"`
    } `json:"bids"`
}

func (s *RateServer) fetchRates(ctx context.Context) (*DepthResponse, error) {
    ctx, span := s.tracer.Start(ctx, "FetchGrinexDepth")
    defer span.End()

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://grinex.io/api/v2/depth?market=usdtrub", nil)
    if err != nil {
        return nil, fmt.Errorf("create request failed: %w", err)
    }

    resp, err := s.Client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }

    var depth DepthResponse
    if err := json.NewDecoder(resp.Body).Decode(&depth); err != nil {
        return nil, fmt.Errorf("decode failed: %w", err)
    }

    span.SetAttributes(
        attribute.Int("ask_count", len(depth.Asks)),
        attribute.Int("bid_count", len(depth.Bids)),
    )

    return &depth, nil
}