package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
)

var (
    RequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "requests_total",
            Help: "Total number of requests",
        },
        []string{"method"},
    )
    RequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "request_duration_seconds",
            Help:    "Request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method"},
    )
    RequestErrors = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "request_errors_total",
            Help: "Total number of request errors",
        },
        []string{"method"},
    )
)

func Init() {
    prometheus.MustRegister(RequestsTotal)
    prometheus.MustRegister(RequestDuration)
    prometheus.MustRegister(RequestErrors)
}