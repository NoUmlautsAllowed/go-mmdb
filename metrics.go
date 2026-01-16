package mmdb

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mmdb_http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"path", "method", "status"},
	)

	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mmdb_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)

	LookupTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mmdb_lookup_total",
			Help: "Total number of IP lookups.",
		},
		[]string{"type"}, // "city", "asn", etc.
	)

	DownloadTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mmdb_download_total",
			Help: "Total number of database downloads.",
		},
		[]string{"database", "status"}, // status: "success", "failure", "skipped"
	)
)
