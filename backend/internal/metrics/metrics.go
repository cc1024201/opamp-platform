package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics OpAMP Platform 监控指标
type Metrics struct {
	// HTTP 请求指标
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestSize     *prometheus.HistogramVec
	HTTPResponseSize    *prometheus.HistogramVec

	// Agent 指标
	AgentsTotal          prometheus.Gauge
	AgentsConnected      prometheus.Gauge
	AgentsDisconnected   prometheus.Gauge
	AgentConnectTotal    prometheus.Counter
	AgentDisconnectTotal prometheus.Counter

	// Configuration 指标
	ConfigurationsTotal      prometheus.Gauge
	ConfigurationChangesTotal prometheus.Counter
	ConfigurationPushTotal    *prometheus.CounterVec

	// 数据库指标
	DBConnectionsOpen prometheus.Gauge
	DBConnectionsIdle prometheus.Gauge
	DBQueriesTotal    *prometheus.CounterVec
	DBQueryDuration   *prometheus.HistogramVec
}

// NewMetrics 创建新的 Metrics 实例
func NewMetrics(namespace string) *Metrics {
	m := &Metrics{
		// HTTP 请求指标
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		HTTPRequestSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_size_bytes",
				Help:      "HTTP request size in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 7),
			},
			[]string{"method", "path"},
		),
		HTTPResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_response_size_bytes",
				Help:      "HTTP response size in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 7),
			},
			[]string{"method", "path"},
		),

		// Agent 指标
		AgentsTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "agents_total",
				Help:      "Total number of agents",
			},
		),
		AgentsConnected: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "agents_connected",
				Help:      "Number of connected agents",
			},
		),
		AgentsDisconnected: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "agents_disconnected",
				Help:      "Number of disconnected agents",
			},
		),
		AgentConnectTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "agent_connect_total",
				Help:      "Total number of agent connections",
			},
		),
		AgentDisconnectTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "agent_disconnect_total",
				Help:      "Total number of agent disconnections",
			},
		),

		// Configuration 指标
		ConfigurationsTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "configurations_total",
				Help:      "Total number of configurations",
			},
		),
		ConfigurationChangesTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "configuration_changes_total",
				Help:      "Total number of configuration changes",
			},
		),
		ConfigurationPushTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "configuration_push_total",
				Help:      "Total number of configuration pushes to agents",
			},
			[]string{"status"},
		),

		// 数据库指标
		DBConnectionsOpen: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_open",
				Help:      "Number of open database connections",
			},
		),
		DBConnectionsIdle: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_idle",
				Help:      "Number of idle database connections",
			},
		),
		DBQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "db_queries_total",
				Help:      "Total number of database queries",
			},
			[]string{"operation"},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "db_query_duration_seconds",
				Help:      "Database query duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"operation"},
		),
	}

	return m
}
