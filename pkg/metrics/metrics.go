// Package metrics provides Prometheus metrics collection for the proxy server.
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/yigitkonur/proxy-http-forward/pkg/config"
)

// Metrics holds all Prometheus metrics for the proxy.
type Metrics struct {
	RequestsTotal     *prometheus.CounterVec
	RequestDuration   *prometheus.HistogramVec
	ActiveConnections prometheus.Gauge
	BytesSent         *prometheus.CounterVec
	BytesReceived     *prometheus.CounterVec
	ErrorsTotal       *prometheus.CounterVec
	TunnelConnections prometheus.Gauge
}

// New creates and registers all metrics.
func New() *Metrics {
	return &Metrics{
		RequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "proxy",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests processed",
			},
			[]string{"method", "status", "type"},
		),
		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "proxy",
				Name:      "request_duration_seconds",
				Help:      "Duration of HTTP request handling in seconds",
				Buckets:   prometheus.ExponentialBuckets(0.001, 2, 15),
			},
			[]string{"method", "type"},
		),
		ActiveConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "proxy",
				Name:      "active_connections",
				Help:      "Number of active client connections",
			},
		),
		BytesSent: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "proxy",
				Name:      "bytes_sent_total",
				Help:      "Total bytes sent to clients",
			},
			[]string{"type"},
		),
		BytesReceived: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "proxy",
				Name:      "bytes_received_total",
				Help:      "Total bytes received from clients",
			},
			[]string{"type"},
		),
		ErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "proxy",
				Name:      "errors_total",
				Help:      "Total number of errors",
			},
			[]string{"type", "reason"},
		),
		TunnelConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "proxy",
				Name:      "tunnel_connections",
				Help:      "Number of active CONNECT tunnel connections",
			},
		),
	}
}

// RecordRequest records a completed request.
func (m *Metrics) RecordRequest(method, status, reqType string, duration float64) {
	m.RequestsTotal.WithLabelValues(method, status, reqType).Inc()
	m.RequestDuration.WithLabelValues(method, reqType).Observe(duration)
}

// RecordError records an error.
func (m *Metrics) RecordError(errType, reason string) {
	m.ErrorsTotal.WithLabelValues(errType, reason).Inc()
}

// IncrementConnections increments active connections counter.
func (m *Metrics) IncrementConnections() {
	m.ActiveConnections.Inc()
}

// DecrementConnections decrements active connections counter.
func (m *Metrics) DecrementConnections() {
	m.ActiveConnections.Dec()
}

// IncrementTunnels increments tunnel connections counter.
func (m *Metrics) IncrementTunnels() {
	m.TunnelConnections.Inc()
}

// DecrementTunnels decrements tunnel connections counter.
func (m *Metrics) DecrementTunnels() {
	m.TunnelConnections.Dec()
}

// Server starts the metrics HTTP server.
type Server struct {
	cfg    config.MetricsConfig
	server *http.Server
}

// NewServer creates a new metrics server.
func NewServer(cfg config.MetricsConfig) *Server {
	mux := http.NewServeMux()
	mux.Handle(cfg.Path, promhttp.Handler())

	return &Server{
		cfg: cfg,
		server: &http.Server{
			Addr:    cfg.Address,
			Handler: mux,
		},
	}
}

// Start starts the metrics server.
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the metrics server.
func (s *Server) Shutdown() error {
	return s.server.Close()
}
