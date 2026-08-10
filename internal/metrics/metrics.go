package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	WorkspaceProvisioningDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "workspace_provisioning_duration_seconds",
			Help:    "Time taken to provision a workspace container.",
			Buckets: []float64{1, 2.5, 5, 10, 15, 20, 30, 45, 60},
		},
	)

	WorkspaceCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "workspace_count",
			Help: "Number of workspaces by status.",
		},
		[]string{"status"},
	)

	ReconciliationLag = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "reconciliation_lag_seconds",
			Help: "Seconds since the last completed reconciliation cycle.",
		},
	)
)

func init() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		WorkspaceProvisioningDuration,
		WorkspaceCount,
		ReconciliationLag,
	)
}
