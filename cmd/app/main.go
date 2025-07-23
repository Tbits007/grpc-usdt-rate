package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"grpc-usdt-rate/internal/config"
	"grpc-usdt-rate/internal/logger"
	_ "grpc-usdt-rate/internal/metrics"
	"grpc-usdt-rate/internal/server"
	"grpc-usdt-rate/internal/storage"
)

func main() {
	cfg := config.Load()
	log := logger.Init()
	defer log.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tp, err := initTracer(ctx, cfg)
	if err != nil {
		log.Fatal("Failed to initialize tracer", zap.Error(err))
	}

	db, err := storage.NewPostgres(cfg.PostgresDSN)
	if err != nil {
		log.Fatal("DB connection failed", zap.Error(err))
	}

	grpcServer := startGRPCServer(cfg, db, log)
	metricsServer := startMetricsServer(cfg, log)

	waitForShutdown(ctx, log, grpcServer, metricsServer, tp, db)
}

func initTracer(ctx context.Context, cfg config.Config) (*sdktrace.TracerProvider, error) {
	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.ServiceName),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

func startGRPCServer(cfg config.Config, db storage.Storage, log *zap.Logger) *grpc.Server {
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal("Failed to listen", zap.String("port", cfg.GRPCPort), zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	srv := server.NewRateServer(db, log)
	srv.Register(grpcServer)

	go func() {
		log.Info("Starting gRPC server", zap.String("address", lis.Addr().String()))
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("gRPC server failed", zap.Error(err))
		}
	}()

	return grpcServer
}

func startMetricsServer(cfg config.Config, log *zap.Logger) *http.Server {
	metricsServer := &http.Server{
		Addr:    ":" + cfg.MetricsPort,
		Handler: promhttp.Handler(),
	}

	go func() {
		log.Info("Starting metrics server", zap.String("address", metricsServer.Addr))
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Metrics server failed", zap.Error(err))
		}
	}()

	return metricsServer
}

func waitForShutdown(
	ctx context.Context,
	log *zap.Logger,
	grpcServer *grpc.Server,
	metricsServer *http.Server,
	tp *sdktrace.TracerProvider,
	db storage.Storage,
) {
	shutdownErr := make(chan error, 1)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		log.Info("Initiating graceful shutdown...")
 
		grpcServer.GracefulStop()

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := metricsServer.Shutdown(ctx); err != nil {
			shutdownErr <- fmt.Errorf("metrics server shutdown error: %w", err)
			return
		}

		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			shutdownErr <- fmt.Errorf("tracer provider shutdown error: %w", err)
			return
		}

		db.Close()

		shutdownErr <- nil
	}()

	if err := <-shutdownErr; err != nil {
		log.Error("Graceful shutdown failed", zap.Error(err))
		os.Exit(1)
	}
	log.Info("All components stopped gracefully")
}