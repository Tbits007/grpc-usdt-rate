package config

import (
	"flag"
	"os"
)

type Config struct {
	PostgresDSN   string
	GRPCPort      string
	MetricsPort   string
	OTLPEndpoint  string
	ServiceName   string
}

func Load() Config {
	cfg := Config{}

	flag.StringVar(&cfg.PostgresDSN, "postgres-dsn", getEnv("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"), "PostgreSQL connection string")
	flag.StringVar(&cfg.GRPCPort, "grpc-port", getEnv("GRPC_PORT", "50051"), "gRPC server port")
	flag.StringVar(&cfg.MetricsPort, "metrics-port", getEnv("METRICS_PORT", "2112"), "Metrics server port")
	flag.StringVar(&cfg.OTLPEndpoint, "otlp-endpoint", getEnv("OTLP_ENDPOINT", "otel-collector:4317"), "OTLP collector endpoint")
	flag.StringVar(&cfg.ServiceName, "service-name", getEnv("SERVICE_NAME", "usdt-rate-service"), "Service name for tracing")

	flag.Parse()

	return cfg
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
