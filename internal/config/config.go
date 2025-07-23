package config

import "os"

type Config struct {
	PostgresDSN         string
	GRPCPort            string
	MetricsPort         string
	OTLPEndpoint        string
	ServiceName         string
}

func Load() Config {
	return Config{
		PostgresDSN:    getEnv("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"),
		GRPCPort:       getEnv("GRPC_PORT", "50051"),
		MetricsPort:    getEnv("METRICS_PORT", "2112"),
		OTLPEndpoint:  getEnv("OTLP_ENDPOINT", "otel-collector:4317"),
		ServiceName:   getEnv("SERVICE_NAME", "usdt-rate-service"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}