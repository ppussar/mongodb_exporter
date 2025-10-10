package internal

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// QueryDuration tracks how long MongoDB queries take
	QueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "mongodb_exporter_query_duration_seconds",
			Help: "Duration of MongoDB queries in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"metric_name", "db", "collection", "query_type"},
	)

	// QueryErrors tracks MongoDB query errors
	QueryErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mongodb_exporter_query_errors_total",
			Help: "Total number of MongoDB query errors",
		},
		[]string{"metric_name", "db", "collection", "error_type"},
	)

	// ActiveQueries tracks currently running queries
	ActiveQueries = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mongodb_exporter_active_queries",
			Help: "Number of currently active MongoDB queries",
		},
		[]string{"db", "collection"},
	)

	// ConnectionStatus tracks MongoDB connection status
	ConnectionStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mongodb_exporter_connection_status",
			Help: "MongoDB connection status (1=connected, 0=disconnected)",
		},
		[]string{"uri"},
	)

	// MetricsCollected tracks successful metric collections
	MetricsCollected = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mongodb_exporter_metrics_collected_total",
			Help: "Total number of metrics successfully collected",
		},
		[]string{"metric_name"},
	)
)
